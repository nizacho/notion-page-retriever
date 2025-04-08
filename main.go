package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jomei/notionapi"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

var blockChildren = make(map[notionapi.BlockID][]notionapi.Block)

// インデントを使ってブロックの階層を視覚的に表現するための補助関数
func getIndent(depth int) string {
	return strings.Repeat("  ", depth)
}

// NotionのページIDを正しいUUIDフォーマットに変換する
func formatPageID(id string) string {
	// すでに正しいフォーマットの場合はそのまま返す
	if strings.Contains(id, "-") {
		return id
	}

	// 32文字のIDを8-4-4-4-12の形式に変換
	if len(id) == 32 {
		return fmt.Sprintf("%s-%s-%s-%s-%s",
			id[0:8],
			id[8:12],
			id[12:16],
			id[16:20],
			id[20:32])
	}

	return id
}

// getRichTextContent combines multiple rich text blocks into a single string
func getRichTextContent(richText []notionapi.RichText) string {
	var content []string
	for _, text := range richText {
		content = append(content, text.PlainText)
	}
	return strings.Join(content, "")
}

// printBlock prints a single block in Notion-like format
func printBlock(block notionapi.Block, depth int) {
	indent := strings.Repeat("    ", depth) // 4スペースでインデント

	switch b := block.(type) {
	case *notionapi.ParagraphBlock:
		fmt.Printf("%s%s\n\n", indent, getRichTextContent(b.Paragraph.RichText))

	case *notionapi.Heading1Block:
		fmt.Printf("%s# %s\n\n", indent, getRichTextContent(b.Heading1.RichText))

	case *notionapi.Heading2Block:
		fmt.Printf("%s## %s\n\n", indent, getRichTextContent(b.Heading2.RichText))

	case *notionapi.Heading3Block:
		fmt.Printf("%s### %s\n\n", indent, getRichTextContent(b.Heading3.RichText))

	case *notionapi.BulletedListItemBlock:
		fmt.Printf("%s- %s\n", indent, getRichTextContent(b.BulletedListItem.RichText))

	case *notionapi.NumberedListItemBlock:
		fmt.Printf("%s1. %s\n", indent, getRichTextContent(b.NumberedListItem.RichText))

	case *notionapi.ToDoBlock:
		checkbox := "[ ]"
		if b.ToDo.Checked {
			checkbox = "[x]"
		}
		fmt.Printf("%s- %s %s\n", indent, checkbox, getRichTextContent(b.ToDo.RichText))

	case *notionapi.ImageBlock:
		if b.Image.Type == "external" {
			fmt.Printf("%s![Image](%s)\n\n", indent, b.Image.External.URL)
		} else if b.Image.Type == "file" {
			fmt.Printf("%s![Image](%s)\n\n", indent, b.Image.File.URL)
		}

	case *notionapi.CodeBlock:
		fmt.Printf("%s```%s\n", indent, b.Code.Language)
		fmt.Printf("%s%s\n", indent, getRichTextContent(b.Code.RichText))
		fmt.Printf("%s```\n\n", indent)

	case *notionapi.QuoteBlock:
		lines := strings.Split(getRichTextContent(b.Quote.RichText), "\n")
		for _, line := range lines {
			fmt.Printf("%s> %s\n", indent, line)
		}
		fmt.Println()

	case *notionapi.CalloutBlock:
		icon := "💡"
		if b.Callout.Icon != nil && b.Callout.Icon.Type == "emoji" {
			icon = string(*b.Callout.Icon.Emoji)
		}
		fmt.Printf("%s> %s %s\n\n", indent, icon, getRichTextContent(b.Callout.RichText))

	case *notionapi.DividerBlock:
		fmt.Printf("%s---\n\n", indent)

	case *notionapi.ToggleBlock:
		fmt.Printf("%s- %s\n", indent, getRichTextContent(b.Toggle.RichText))

	case *notionapi.TableBlock:
		// テーブルヘッダーとデータは子ブロックとして取得されるため、
		// ここでは何も出力せず、子ブロックの処理に任せる
		return

	case *notionapi.TableRowBlock:
		cells := []string{}
		for _, cell := range b.TableRow.Cells {
			cells = append(cells, getRichTextContent(cell))
		}
		fmt.Printf("%s| %s |\n", indent, strings.Join(cells, " | "))

	case *notionapi.ColumnListBlock, *notionapi.ColumnBlock:
		// カラムブロックは視覚的な構造のみなので、
		// 内容は子ブロックとして処理される
		return
	}
}

func summarizeContent(content string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	resp, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("あなたは与えられたテキストを要約する専門家です。重要なポイントを箇条書きで3-5個程度にまとめてください。"),
				openai.UserMessage(content),
			},
			Model: shared.ChatModelGPT4,
		},
	)

	if err != nil {
		return "", fmt.Errorf("summarization failed: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <page-id>")
		os.Exit(1)
	}

	token := os.Getenv("NOTION_API_TOKEN")
	if token == "" {
		log.Fatal("NOTION_API_TOKEN is not set")
	}

	pageID := formatPageID(os.Args[1])
	client := notionapi.NewClient(notionapi.Token(token))

	// 表示用の出力
	err := printBlocksRecursive(client, notionapi.BlockID(pageID), 0, nil)
	if err != nil {
		log.Fatalf("Error fetching blocks: %v", err)
	}

	// 要約用のテキスト収集
	var contentBuilder strings.Builder
	err = collectContent(client, notionapi.BlockID(pageID), &contentBuilder)
	if err != nil {
		log.Fatalf("Error collecting content: %v", err)
	}

	content := contentBuilder.String()
	fmt.Println("\n=== AI による要約 ===\n")
	summary, err := summarizeContent(content)
	if err != nil {
		log.Printf("Error generating summary: %v", err)
	} else {
		fmt.Println(summary)
	}
}

// fetchChildBlocks は指定されたブロックIDの子ブロックを再帰的に取得します
func fetchChildBlocks(ctx context.Context, blockID notionapi.BlockID, client *notionapi.Client) ([]notionapi.Block, error) {
	var blocks []notionapi.Block
	var cursor notionapi.Cursor

	for {
		// ページネーションを使用してブロックを取得
		resp, err := client.Block.GetChildren(ctx, blockID, &notionapi.Pagination{
			StartCursor: cursor,
			PageSize:    100,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get blocks: %v", err)
		}

		blocks = append(blocks, resp.Results...)

		// 次のページがない場合は終了
		if !resp.HasMore {
			break
		}

		// 次のページのカーソルを設定
		cursor = notionapi.Cursor(resp.NextCursor)
	}

	// 各ブロックの子ブロックを再帰的に取得
	for _, block := range blocks {
		if block.GetHasChildren() {
			childBlocks, err := fetchChildBlocks(ctx, block.GetID(), client)
			if err != nil {
				return nil, err
			}
			// Store child blocks in the map
			blockChildren[block.GetID()] = childBlocks
		}
	}

	return blocks, nil
}

// printBlocksRecursive prints blocks recursively with proper indentation
func printBlocksRecursive(client *notionapi.Client, blockID notionapi.BlockID, depth int, contentBuilder *strings.Builder) error {
	blocks, err := fetchChildBlocks(context.Background(), blockID, client)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		printBlock(block, depth)
		if _, ok := blockChildren[block.GetID()]; ok {
			err := printBlocksRecursive(client, block.GetID(), depth+1, contentBuilder)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// collectContent collects text content from blocks for summarization
func collectContent(client *notionapi.Client, blockID notionapi.BlockID, contentBuilder *strings.Builder) error {
	blocks, err := fetchChildBlocks(context.Background(), blockID, client)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			contentBuilder.WriteString(getRichTextContent(b.Paragraph.RichText))
		case *notionapi.Heading1Block:
			contentBuilder.WriteString(getRichTextContent(b.Heading1.RichText))
		case *notionapi.Heading2Block:
			contentBuilder.WriteString(getRichTextContent(b.Heading2.RichText))
		case *notionapi.Heading3Block:
			contentBuilder.WriteString(getRichTextContent(b.Heading3.RichText))
		case *notionapi.BulletedListItemBlock:
			contentBuilder.WriteString(getRichTextContent(b.BulletedListItem.RichText))
		case *notionapi.NumberedListItemBlock:
			contentBuilder.WriteString(getRichTextContent(b.NumberedListItem.RichText))
		case *notionapi.ToDoBlock:
			contentBuilder.WriteString(getRichTextContent(b.ToDo.RichText))
		case *notionapi.QuoteBlock:
			contentBuilder.WriteString(getRichTextContent(b.Quote.RichText))
		case *notionapi.CalloutBlock:
			contentBuilder.WriteString(getRichTextContent(b.Callout.RichText))
		case *notionapi.ToggleBlock:
			contentBuilder.WriteString(getRichTextContent(b.Toggle.RichText))
		}
		contentBuilder.WriteString("\n\n")
		if _, ok := blockChildren[block.GetID()]; ok {
			err := collectContent(client, block.GetID(), contentBuilder)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
