#!/usr/bin/env bash
# LocalStack の ready hook。bucket 作成 + 匿名 read policy 付与。
# browser から `http://localhost:4566/<bucket>/<key>` を直接開いて画像を表示できるようにする。
set -euo pipefail

BUCKET="${S3_BUCKET_NAME:-kakkky-dev-images}"

awslocal s3api create-bucket --bucket "$BUCKET" >/dev/null 2>&1 || true

awslocal s3api put-bucket-policy --bucket "$BUCKET" --policy "$(cat <<JSON
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicRead",
      "Effect": "Allow",
      "Principal": "*",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::${BUCKET}/*"]
    }
  ]
}
JSON
)"

echo "localstack: bucket '${BUCKET}' ready"
