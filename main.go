package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jomei/notionapi"
)

var client *notionapi.Client
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

func main() {
	// Get the Notion API token from environment variable
	token := os.Getenv("NOTION_API_TOKEN")
	if token == "" {
		log.Fatal("NOTION_API_TOKEN environment variable is not set")
	}

	// Initialize the Notion client
	client = notionapi.NewClient(notionapi.Token(token))

	// Get the page ID from command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Please provide a page ID as an argument")
	}
	pageID := formatPageID(os.Args[1])

	// Create context
	ctx := context.Background()

	// Fetch and print blocks
	blocks, err := fetchChildBlocks(ctx, notionapi.BlockID(pageID))
	if err != nil {
		log.Fatalf("Error fetching blocks: %v", err)
	}

	// Print blocks with indentation
	printBlocksRecursive(blocks, 0)
}

// fetchChildBlocks は指定されたブロックIDの子ブロックを再帰的に取得します
func fetchChildBlocks(ctx context.Context, blockID notionapi.BlockID) ([]notionapi.Block, error) {
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
			childBlocks, err := fetchChildBlocks(ctx, block.GetID())
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
func printBlocksRecursive(blocks []notionapi.Block, depth int) {
	for _, block := range blocks {
		printBlock(block, depth)
		if children, ok := blockChildren[block.GetID()]; ok {
			printBlocksRecursive(children, depth+1)
		}
	}
}
