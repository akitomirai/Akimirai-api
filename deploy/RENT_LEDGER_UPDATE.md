# Rent Ledger 远程更新发布说明

App 会检查：

```text
https://akimirai.xyz/rent-ledger/latest.json
```

清单示例在 `deploy/rent-ledger-updates/latest.json`。`apkUrl` 可以写相对路径，例如 `rent-ledger-app.apk`，App 会解析为：

```text
https://akimirai.xyz/rent-ledger/rent-ledger-app.apk
```

## Caddy

参考 `deploy/rent-ledger-update.Caddyfile.example`，给 `akimirai.xyz` 增加静态目录：

```caddy
handle_path /rent-ledger/* {
    root * /var/www/rent-ledger
    file_server
    header Cache-Control "no-store"
}
```

## 发布步骤

1. 构建新版 APK，并保证 `versionCode` 大于当前安装版本。
2. 上传 APK 到服务器目录：

```bash
mkdir -p /var/www/rent-ledger
cp app-release.apk /var/www/rent-ledger/rent-ledger-app.apk
```

3. 更新 `/var/www/rent-ledger/latest.json`：

```json
{
  "versionCode": 2,
  "versionName": "1.1",
  "apkUrl": "rent-ledger-app.apk",
  "notes": "更新说明",
  "mandatory": false,
  "sha256": ""
}
```

4. 可选：填写 APK 的 SHA-256。填写后 App 会校验下载文件。

```bash
sha256sum /var/www/rent-ledger/rent-ledger-app.apk
```

5. 确认清单可访问：

```bash
curl https://akimirai.xyz/rent-ledger/latest.json
```
