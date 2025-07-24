# 🤝 Contributing to StreamForge

StreamForgeプロジェクトへのコントリビューションをお考えいただき、ありがとうございます！このガイドでは、効果的にコントリビューションするための手順とガイドラインを説明します。

## 📋 Table of Contents

- [🎯 コントリビューションの種類](#-コントリビューションの種類)
- [🚀 開発環境セットアップ](#-開発環境セットアップ)
- [📝 コーディング規約](#-コーディング規約)
- [🧪 テスト戦略](#-テスト戦略)
- [📋 PR (Pull Request) プロセス](#-pr-pull-request-プロセス)
- [🔍 Issue 報告](#-issue-報告)
- [📚 ドキュメント改善](#-ドキュメント改善)
- [🏆 コミュニティガイドライン](#-コミュニティガイドライン)

## 🎯 コントリビューションの種類

StreamForgeでは以下のような貢献を歓迎しています：

### コード貢献
- 🐛 **バグ修正**: 既存のバグレポートの修正
- ✨ **新機能**: 新しい機能の実装
- ⚡ **パフォーマンス改善**: 処理速度やメモリ使用量の最適化
- 🎨 **リファクタリング**: コード品質向上のための構造改善

### 非コード貢献
- 📖 **ドキュメント**: README、API、チュートリアルの改善
- 🧪 **テスト**: テストケースの追加・改善
- 🎨 **デザイン**: UI/UX の改善提案
- 🌐 **翻訳**: 国際化対応のための翻訳作業

### コミュニティ貢献
- 🐛 **バグ報告**: 詳細なバグレポートの作成
- 💡 **機能提案**: 新機能のアイデア提案
- 📝 **Issue トリアージ**: Issueの分類・優先順位付け
- 💬 **サポート**: ディスカッションでの質問回答

## 🚀 開発環境セットアップ

### 前提条件

以下のツールがインストール済みである必要があります：

```bash
# バージョン確認
node --version  # v20.0.0 以上
pnpm --version  # v8.0.0 以上
go version      # v1.21 以上
rustc --version # v1.75.0 以上
python --version # v3.11 以上
docker --version # v24.0.0 以上
kubectl version  # v1.25 以上
```

### 環境構築手順

1. **リポジトリクローン**
```bash
git clone https://github.com/streamforge/streamforge.git
cd streamforge
```

2. **依存関係インストール**
```bash
# 全言語の依存関係を一括インストール
make dev-setup

# または個別インストール
pnpm install                    # Node.js dependencies
cargo build                     # Rust dependencies
go mod download                 # Go dependencies  
pip install -r requirements.txt # Python dependencies
```

3. **開発サーバー起動**
```bash
# Docker Compose で開発環境全体を起動
make dev-up

# または個別起動
make dev-dashboard      # Next.js dashboard
make dev-api-gateway    # Go API gateway
make dev-stream-processor # Rust stream processor
make dev-ml-engine      # Python ML engine
```

4. **動作確認**
```bash
# ヘルスチェック
make health-check

# テスト実行
make test-all

# リンター実行
make lint-all
```

### Devcontainer対応

VSCode + Devcontainer環境も用意しています：

1. VSCode で `streamforge` フォルダを開く
2. `Reopen in Container` を選択
3. 自動で開発環境が構築されます

## 📝 コーディング規約

### 共通規約

- **命名規則**: 英語で具体的かつ意味のある名前を使用
- **コメント**: 複雑なロジックや意図が不明瞭な箇所のみ
- **フォーマット**: 各言語の標準フォーマッターを使用
- **セキュリティ**: 機密情報をコードに含めない

### 言語別規約

#### Go
```bash
# フォーマット・リンター
make fmt-go
make lint-go

# ファイル構成
pkg/                 # 外部公開パッケージ
internal/           # 内部パッケージ
cmd/                # エントリーポイント
```

#### Rust
```bash
# フォーマット・リンター
make fmt-rust
make lint-rust

# Cargo.toml設定
[dependencies]
serde = { version = "1.0", features = ["derive"] }
tokio = { version = "1.0", features = ["full"] }
```

#### TypeScript
```bash
# フォーマット・リンター
make fmt-ts
make lint-ts

# tsconfig.json 設定
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true
  }
}
```

#### Python
```bash
# フォーマット・リンター
make fmt-python
make lint-python

# pyproject.toml 設定
[tool.ruff]
select = ["E", "F", "I", "B", "SIM"]
line-length = 100
```

## 🧪 テスト戦略

### テストレベル

1. **Unit Tests** (カバレッジ: 90%以上)
```bash
make test-unit           # 全言語のユニットテスト
make test-unit-go        # Go
make test-unit-rust      # Rust  
make test-unit-ts        # TypeScript
make test-unit-python    # Python
```

2. **Integration Tests**
```bash
make test-integration    # API間結合テスト
```

3. **E2E Tests**
```bash
make test-e2e           # Playwright E2Eテスト
```

4. **Performance Tests**
```bash
make test-benchmark     # パフォーマンステスト
```

### テスト作成ガイドライン

- **AAA パターン**: Arrange, Act, Assert の順序
- **命名規則**: `Test<機能>_<条件>_<期待結果>`
- **テストデータ**: fixtures/ フォルダで管理
- **モック**: 外部依存は必ずモック化

## 📋 PR (Pull Request) プロセス

### PR作成前チェックリスト

- [ ] ブランチ名が規約に従っている (`feature/xxx`, `fix/xxx`, `docs/xxx`)
- [ ] コミットメッセージがConventional Commits形式
- [ ] 全テストが通過している
- [ ] リンターエラーがない
- [ ] 関連Issueが存在する

### PR作成手順

1. **ブランチ作成**
```bash
git checkout -b feature/add-anomaly-detection
```

2. **開発・テスト**
```bash
# 開発作業
# ...

# テスト実行
make test-all
make lint-all
```

3. **コミット**
```bash
git add .
git commit -m "feat: AI駆動異常検知エンジンを追加

- 機械学習モデルによるリアルタイム異常検知
- 根本原因分析機能を実装
- Prometheus メトリクス統合

Closes #123"
```

4. **プッシュ・PR作成**
```bash
git push origin feature/add-anomaly-detection
# GitHub上でPR作成
```

### PR レビュープロセス

1. **自動チェック**: CI/CDパイプラインによる自動テスト
2. **コードレビュー**: 最低2名の承認が必要
3. **セキュリティレビュー**: セキュリティチームによる確認
4. **マージ**: Squash and Merge で統合

### コミットメッセージ規約

[Conventional Commits](https://www.conventionalcommits.org/) 形式を採用：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Type例:**
- `feat`: 新機能
- `fix`: バグ修正
- `docs`: ドキュメント更新
- `style`: フォーマット変更
- `refactor`: リファクタリング
- `test`: テスト追加・修正
- `chore`: ビルド・補助ツール変更

## 🔍 Issue 報告

### バグ報告

バグを発見した場合は、以下のテンプレートを使用してください：

```markdown
## 🐛 バグ概要
簡潔にバグの内容を説明してください。

## 🔄 再現手順
1. '...' に移動
2. '...' をクリック
3. '...' まで下にスクロール
4. エラーを確認

## ✅ 期待される動作
何が起こることを期待していたかを説明してください。

## ❌ 実際の動作
実際に何が起こったかを説明してください。

## 📷 スクリーンショット
可能であれば、スクリーンショットを追加してください。

## 🖥️ 環境情報
- OS: [例: macOS 14.0]
- Browser: [例: Chrome 120.0]
- Version: [例: v1.2.3]

## 📝 追加情報
他に関連する情報があれば記載してください。
```

### 機能要求

新機能の提案は以下のテンプレートを使用：

```markdown
## 💡 機能概要
提案する機能の概要を説明してください。

## 🎯 解決したい課題
この機能がどのような課題を解決するかを説明してください。

## 💭 提案する解決策
理想的な解決策を詳しく説明してください。

## 🔄 代替案
検討した代替案があれば説明してください。

## 📋 受け入れ基準
機能が完成したと判断する基準を列挙してください。

- [ ] 基準1
- [ ] 基準2
- [ ] 基準3
```

## 📚 ドキュメント改善

### ドキュメント種別

1. **API ドキュメント**: OpenAPI/GraphQL schema から自動生成
2. **ユーザーガイド**: `docs/` フォルダのMarkdown
3. **開発者ガイド**: `docs/development/` フォルダ
4. **アーキテクチャ**: `docs/architecture/` と ADR

### ドキュメント更新手順

1. **Markdown編集**
```bash
# ドキュメントサイト起動
cd docs/website
pnpm dev
```

2. **プレビュー確認**
```bash
# ローカルでプレビュー
open http://localhost:3000
```

3. **自動生成更新**
```bash
# API ドキュメント更新
make docs-generate-api

# コード例更新  
make docs-generate-examples
```

## 🏆 コミュニティガイドライン

### 行動規範

すべての参加者は [Code of Conduct](./CODE_OF_CONDUCT.md) に従う必要があります。

### コミュニケーション

- **GitHub Discussions**: 質問・提案・雑談
- **GitHub Issues**: バグ報告・機能要求
- **Discord**: リアルタイムチャット (招待制)

### 初回コントリビューター向け

`good first issue` ラベルが付いたIssueから始めることをお勧めします：

- ドキュメント誤字修正
- 簡単なバグ修正
- テストケース追加
- サンプルコード改善

### 表彰制度

定期的に以下の表彰を行います：

- **🥇 Contributor of the Month**: 月間最優秀コントリビューター
- **🏆 Security Champion**: セキュリティ改善への貢献
- **📚 Documentation Hero**: ドキュメント改善への貢献
- **🧪 Testing Master**: テスト品質向上への貢献

## 📞 サポート

困ったときは遠慮なくお声がけください：

- **GitHub Discussions**: [General Q&A](https://github.com/streamforge/streamforge/discussions)
- **Discord**: [StreamForge Community](https://discord.gg/streamforge) (招待制)
- **Email**: contribute@streamforge.io

---

**🙏 Thank you for contributing to StreamForge!** 