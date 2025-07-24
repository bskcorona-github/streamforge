# StreamForge TypeScript SDK

StreamForgeのTypeScript SDKです。メトリクスの送信、クエリ、アラート管理などの機能を提供します。

## インストール

```bash
npm install @streamforge/sdk-typescript
```

## 使用方法

### 基本的な使用例

```typescript
import { StreamForgeClient } from '@streamforge/sdk-typescript';

// クライアントの初期化
const client = new StreamForgeClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-api-key',
});

// メトリクスの送信
const metrics = [
  {
    name: 'cpu_usage',
    value: 75.5,
    timestamp: Date.now(),
    type: 'gauge',
    labels: {
      host: 'server-01',
      region: 'us-west-1'
    }
  }
];

await client.sendMetrics(metrics);

// メトリクスのクエリ
const result = await client.queryMetrics({
  startTime: Date.now() - 3600000, // 1時間前
  endTime: Date.now(),
  limit: 100,
  filters: {
    name: 'cpu_usage'
  }
});

console.log(result.data);
```

### アラートルールの管理

```typescript
// アラートルールの作成
const rule = await client.createAlertRule({
  name: 'High CPU Usage',
  description: 'Alert when CPU usage exceeds 90%',
  query: 'cpu_usage > 90',
  condition: {
    operator: '>',
    threshold: 90,
    duration: '5m'
  },
  severity: 'high',
  notifications: [
    {
      type: 'email',
      config: {
        to: 'admin@example.com'
      }
    }
  ]
});

// アラートルールの取得
const rules = await client.getAlertRules();
console.log(rules);

// アラートルールの削除
await client.deleteAlertRule(rule.id);
```

### エラーハンドリング

```typescript
import { StreamForgeError, ValidationError } from '@streamforge/sdk-typescript';

try {
  await client.sendMetrics(invalidMetrics);
} catch (error) {
  if (error instanceof ValidationError) {
    console.error('Validation error:', error.message);
  } else if (error instanceof StreamForgeError) {
    console.error('StreamForge error:', error.message, error.code);
  } else {
    console.error('Unexpected error:', error);
  }
}
```

## API リファレンス

### StreamForgeClient

#### コンストラクタ

```typescript
new StreamForgeClient(config: StreamForgeConfig)
```

#### 設定オプション

- `baseURL`: APIのベースURL（デフォルト: `http://localhost:8080`）
- `timeout`: リクエストタイムアウト（ミリ秒、デフォルト: `30000`）
- `retries`: リトライ回数（デフォルト: `3`）
- `apiKey`: API認証キー

#### メソッド

##### sendMetrics(data: MetricData[]): Promise<void>

メトリクスデータを送信します。

##### queryMetrics(options: QueryOptions): Promise<QueryResult>

メトリクスデータをクエリします。

##### getHealth(): Promise<HealthStatus>

システムの健全性を取得します。

##### getSystemMetrics(): Promise<SystemMetrics>

システムメトリクスを取得します。

##### createAlertRule(rule: AlertRule): Promise<AlertRule>

アラートルールを作成します。

##### getAlertRules(): Promise<AlertRule[]>

アラートルール一覧を取得します。

##### deleteAlertRule(ruleId: string): Promise<void>

アラートルールを削除します。

### 型定義

詳細な型定義については、`src/types.ts`を参照してください。

## 開発

### セットアップ

```bash
npm install
```

### ビルド

```bash
npm run build
```

### テスト

```bash
npm test
```

### リント

```bash
npm run lint
```

## ライセンス

Apache-2.0 