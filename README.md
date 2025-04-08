# Notion Page Structure Retriever

Notionページの階層構造を再帰的に取得し、Markdown形式で出力するコマンドラインツールです。
ページの内容をAIで要約する機能も備えています。

## 特徴

- Notion APIを使用してページの階層構造を再帰的に取得
- 深さ優先探索（DFS）でブロックを取得
- ページネーション対応で大きなページも取得可能
- 階層構造を視覚的に表現（インデント）
- OpenAI GPT-4を使用したページ内容の要約機能

## サポートされているブロック

- 見出し (H1, H2, H3)
- 段落
- リスト（箇条書き、番号付き）
- チェックボックス
- 画像
- コードブロック
- 引用
- コールアウト
- 区切り線
- トグル
- テーブル

## 前提条件

- Go 1.22以上
- Notion APIトークン
- OpenAI APIキー
- アクセス権のあるNotionページID

## インストール

```bash
git clone https://github.com/nizacho/notion-page-retriever.git
cd notion-page-retriever
go mod download
```

## 使用方法

1. Notion APIトークンを環境変数に設定：
```bash
export NOTION_API_TOKEN="your-notion-api-token"
```

2. OpenAI APIキーを環境変数に設定：
```bash
export OPENAI_API_KEY="your-openai-api-key"
```

3. プログラムを実行：
```bash
go run main.go <page-id>
```

### ページIDの取得方法

NotionのページURLから取得できます：
- `https://www.notion.so/myworkspace/My-Page-1ba1af0e3602808ea8ddfbeb8c0b6071`
  → `1ba1af0e3602808ea8ddfbeb8c0b6071`

## 出力形式

プログラムは以下の2つの部分で構成された出力を生成します：

1. ページの階層構造（Markdown形式）
   - 4スペースでインデント
   - Markdownの見出し記法（#, ##, ###）
   - リスト記法（-, 1.）
   - その他のNotionブロックをMarkdown形式で表現

2. AIによる要約
   - 区切り線 `=== AI による要約 ===`
   - 重要なポイントを箇条書きで3-5個程度に要約

### 出力例

```markdown
# 見出し1

## 見出し2
    - 箇条書き1
    - 箇条書き2
        1. 番号付きリスト1
        2. 番号付きリスト2

=== AI による要約 ===

- 重要なポイント1
- 重要なポイント2
- 重要なポイント3
```

## トラブルシューティング

1. APIトークンが認識されない
   - 環境変数が正しく設定されているか確認
   - トークンの形式が正しいか確認

2. ページにアクセスできない
   - ページIDが正しいか確認
   - APIトークンにページへのアクセス権があるか確認

## ライセンス

MIT License

## 謝辞

このプロジェクトは以下のライブラリを使用しています：
- [github.com/jomei/notionapi](https://github.com/jomei/notionapi)
- [github.com/openai/openai-go](https://github.com/openai/openai-go) (ベータ版)

## 注意事項

### OpenAI Go公式ライブラリについて
このプロジェクトでは、OpenAIの公式Go言語ライブラリ（`openai-go`）を使用していますが、このライブラリは現在ベータ版であり、正式にリリースされていません。そのため、APIの仕様が変更される可能性があります。

### AIプロンプトについて
このプロジェクトでは、以下のデフォルト設定でAIによる要約を行っています：

- **使用モデル**: GPT-4
- **システムプロンプト**: "あなたは与えられたテキストを要約する専門家です。重要なポイントを箇条書きで3-5個程度にまとめてください。"

これらの設定は`main.go`の`summarizeContent`関数内でハードコードされており、現時点ではコマンドライン引数などによる動的な変更はサポートしていません。必要に応じてソースコードを修正してください。 