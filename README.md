# 🔥 StreamForge

[![CI](https://github.com/streamforge/streamforge/workflows/CI/badge.svg)](https://github.com/streamforge/streamforge/actions)
[![codecov](https://codecov.io/gh/streamforge/streamforge/branch/main/graph/badge.svg)](https://codecov.io/gh/streamforge/streamforge)
[![Security](https://github.com/streamforge/streamforge/workflows/Security/badge.svg)](https://github.com/streamforge/streamforge/actions)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/streamforge/streamforge/badge)](https://securityscorecards.dev/viewer/?uri=github.com/streamforge/streamforge)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/streamforge/streamforge)](https://goreportcard.com/report/github.com/streamforge/streamforge)
[![Rust](https://img.shields.io/badge/rust-1.75%2B-orange.svg)](https://www.rust-lang.org/)
[![TypeScript](https://img.shields.io/badge/typescript-5.3%2B-blue.svg)](https://www.typescriptlang.org/)
[![Python](https://img.shields.io/badge/python-3.11%2B-green.svg)](https://www.python.org/)

> 🚀 **リアルタイム分散システムの観測・監視・分析を統合する次世代オブザーバビリティプラットフォーム**

## ✨ 特徴

- **🔄 統合データ収集**: OpenTelemetryベースの包括的なオブザーバビリティデータ収集
- **⚡ 高速ストリーム処理**: Rust製エンジンによる低レイテンシリアルタイム分析
- **🤖 AI駆動異常検知**: 機械学習による自動的な異常パターン検出と根本原因分析
- **📊 直感的ダッシュボード**: Next.js + React Server Componentsによるモダンな可視化UI
- **☸️ Kubernetes ネイティブ**: Operatorパターンによる自動運用とスケーリング
- **🌍 マルチクラウド対応**: AWS、GCP、Azure、オンプレミス環境での統一運用

## 🚀 クイックスタート

### 前提条件

- Docker & Docker Compose
- Kubernetes 1.25+ (開発時はk3d/kind推奨)
- Node.js 20+ & pnpm
- Rust 1.75+ & Cargo
- Go 1.21+
- Python 3.11+

### 1. 開発環境構築

```bash
# リポジトリクローン
git clone https://github.com/streamforge/streamforge.git
cd streamforge

# 開発環境セットアップ (Devcontainer対応)
make dev-setup

# 全コンポーネント起動
make dev-up
```

### 2. デモ環境体験

```bash
# デモデータ生成付きで起動
make demo

# ダッシュボードアクセス
open http://localhost:3000
```

### 3. Kubernetes デプロイ

```bash
# Helm チャートでインストール
helm repo add streamforge https://streamforge.github.io/helm-charts
helm install streamforge streamforge/streamforge --namespace observability --create-namespace
```

## 📋 プロジェクト構成

```
streamforge/
├── apps/                          # アプリケーション群
│   ├── dashboard/                 # Next.js ダッシュボード UI
│   ├── api-gateway/               # GraphQL/REST API ゲートウェイ (Go)
│   ├── stream-processor/          # リアルタイム処理エンジン (Rust)
│   ├── ml-engine/                 # AI異常検知エンジン (Python)
│   ├── collector/                 # データ収集エージェント (Go)
│   └── operator/                  # Kubernetes Operator (Go)
├── packages/                      # 共有ライブラリ
│   ├── proto/                     # Protocol Buffers 定義
│   ├── sdk-typescript/            # TypeScript SDK
│   ├── sdk-go/                    # Go SDK
│   ├── sdk-rust/                  # Rust SDK
│   └── sdk-python/                # Python SDK
├── infra/                         # インフラ定義
│   ├── terraform/                 # Terraform IaC
│   ├── pulumi/                    # Pulumi IaC (代替)
│   ├── helm/                      # Helm Charts
│   ├── k8s/                       # Kubernetes マニフェスト
│   └── docker/                    # Docker 設定
├── docs/                          # ドキュメント
│   ├── website/                   # Docusaurus サイト
│   ├── api/                       # API ドキュメント
│   ├── adr/                       # アーキテクチャ決定記録
│   └── runbooks/                  # 運用手順書
├── examples/                      # サンプル・チュートリアル
├── benchmarks/                    # パフォーマンステスト
├── scripts/                       # 自動化スクリプト
├── tools/                         # 開発ツール
└── .github/                       # GitHub 設定
    ├── workflows/                 # CI/CD パイプライン
    ├── ISSUE_TEMPLATE/            # Issue テンプレート
    └── PULL_REQUEST_TEMPLATE.md   # PR テンプレート
```

## 🛠️ 技術スタック

### Backend
- **Rust**: 低レイテンシストリーム処理、パフォーマンス重要箇所
- **Go**: API サーバー、Kubernetes Operator、収集エージェント
- **Python**: 機械学習・データサイエンス処理

### Frontend
- **Next.js 14**: App Router + React Server Components
- **TypeScript**: 型安全なフロントエンド開発
- **TailwindCSS**: モダンなデザインシステム

### Data & Infrastructure
- **Apache Kafka/Pulsar**: イベントストリーミング
- **ClickHouse**: 高速時系列データベース
- **Redis**: キャッシュ・セッション管理
- **PostgreSQL**: メタデータ管理

### Observability
- **OpenTelemetry**: 統合テレメトリ標準
- **Prometheus**: メトリクス収集
- **Jaeger**: 分散トレーシング
- **Grafana**: 可視化・アラート

### DevOps
- **Kubernetes**: コンテナオーケストレーション
- **Terraform/Pulumi**: Infrastructure as Code
- **ArgoCD**: GitOps デプロイメント
- **GitHub Actions**: CI/CD パイプライン

## 📚 ドキュメント

- [📖 **公式ドキュメント**](https://streamforge.github.io/docs/)
- [🏗️ **アーキテクチャガイド**](./docs/architecture/)
- [🚀 **API リファレンス**](./docs/api/)
- [🎯 **チュートリアル**](./examples/)
- [🔧 **運用ガイド**](./docs/runbooks/)

## 🤝 コントリビューション

StreamForgeはオープンソースプロジェクトです。コントリビューションを歓迎します！

1. [Contributing Guide](./CONTRIBUTING.md) を確認
2. [Code of Conduct](./CODE_OF_CONDUCT.md) に同意
3. Issue作成またはPull Request送信

### 開発者向けリソース

- [開発環境セットアップ](./docs/development/setup.md)
- [コーディング規約](./docs/development/coding-standards.md)
- [テスト戦略](./docs/development/testing.md)
- [リリースプロセス](./docs/development/release.md)

## 📈 ロードマップ

- [x] **Phase 1** (90日): コア収集・処理エンジン ✅
- [ ] **Phase 2** (180日): AI異常検知、アラート管理
- [ ] **Phase 3** (365日): エッジ対応、予測分析

詳細は [Project Roadmap](https://github.com/streamforge/streamforge/projects/1) を参照

## 📄 ライセンス

Apache License 2.0 - 詳細は [LICENSE](./LICENSE) を参照

## 🔐 セキュリティ

セキュリティの脆弱性を発見した場合は、[SECURITY.md](./SECURITY.md) の手順に従って報告してください。

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=streamforge/streamforge&type=Date)](https://star-history.com/#streamforge/streamforge&Date)

---

**Built with ❤️ by the StreamForge community** 