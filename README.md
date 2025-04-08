# notion-page-retriever

このアプリケーションは、NotionのページIDを指定して、そのページの階層構造を再帰的に取得し、Markdownフォーマットで出力するコマンドラインツールです。

## 機能

- Notion APIを使用してページの階層構造を取得
- 深さ優先探索（DFS）による再帰的なブロック取得
- ページネーション対応（100ブロックずつ取得）
- 階層構造を視覚的に表現するインデント
- Notionに貼り付け可能なMarkdown形式での出力

### サポートされているブロックタイプ

- 見出し (H1, H2, H3)
- 段落
- 箇条書きリスト
- 番号付きリスト
- チェックボックス
- 画像（外部URL、ファイル）
- コードブロック（言語指定付き）
- 引用
- コールアウト（絵文字アイコン付き）
- 区切り線
- トグル
- テーブル

## 必要条件

- Go 1.22以上
- Notion API トークン
- アクセス権のあるNotionページのID

## インストール

1. リポジトリをクローン：
```bash
git clone https://github.com/nizacho/notion-page-retriever.git
cd notion-page-retriever
```

2. 依存関係をインストール：
```bash
go mod download
```

## 使い方

1. Notion API トークンを環境変数に設定：
```bash
export NOTION_API_TOKEN="your-notion-api-token"
```

2. プログラムを実行：
```bash
go run main.go <page-id>
```

### ページIDの取得方法

1. NotionでページをURLをコピー
2. 以下の形式のURLからページIDを抽出：
   - `https://www.notion.so/your-workspace/page-title-1ba1af0e3602808ea8ddfbeb8c0b6071`
   - この場合、ページIDは `1ba1af0e3602808ea8ddfbeb8c0b6071`

注：プログラムは自動的にページIDを正しいUUID形式（`1ba1af0e-3602-808e-a8dd-fbeb8c0b6071`）に変換します。

### 出力形式

プログラムは以下のような形式で出力します：

```markdown
# 見出し1

---

- 箇条書き1
    - サブアイテム1
    - サブアイテム2
        - サブサブアイテム
- 箇条書き2
    - [ ] タスク1
    - [x] 完了したタスク

> 引用テキスト

```

## セットアップのトラブルシューティング

1. API トークンが認識されない場合：
   - 環境変数が正しく設定されているか確認
   - トークンが有効であることを確認

2. ページにアクセスできない場合：
   - ページがインテグレーションと共有されているか確認
   - ページIDが正しいか確認
   - 必要な権限が付与されているか確認

## ライセンス

このプロジェクトは[MITライセンス](LICENSE)の下で公開されています。

## 謝辞

このプロジェクトは以下のライブラリを使用しています：
- [github.com/jomei/notionapi](https://github.com/jomei/notionapi) - Notion APIクライアント 