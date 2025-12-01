# s3-deploy-action

deploy website on aliyun OSS(Alibaba Cloud OSS) or Tencent Cloud COS.

将静态网站部署在阿里云 OSS 或腾讯云 COS。

## 概览

- 在阿里云 OSS 或腾讯云 COS 创建一个存放网站的 bucket
- 准备一个域名，可能需要备案
- 在你的网站 repo 中，配置 github action, action 触发则**增量上传**网站 repo 生成的资源文件到 bucket 中
- 通过 CDN, 可以很方便地加速网站的访问，支持 HTTPS

## Usage

### Aliyun OSS (阿里云 OSS)

```yml
    - name: upload files to OSS
      uses: eallion/s3-deploy-action@v1
      with:
          provider: aliyun # Optional, default is aliyun
          accessKeyId: ${{ secrets.ACCESS_KEY_ID }}
          accessKeySecret: ${{ secrets.ACCESS_KEY_SECRET }}
          bucket: your-bucket-name
          # use your own endpoint
          endpoint: oss-cn-shanghai.aliyuncs.com
          folder: your-website-output-folder
```

### Tencent Cloud COS (腾讯云 COS)

```yml
    - name: upload files to COS
      uses: eallion/s3-deploy-action@v1
      with:
          provider: tencent
          cos_secret_id: ${{ secrets.COS_SECRET_ID }}
          cos_secret_key: ${{ secrets.COS_SECRET_KEY }}
          cos_bucket: your-bucket-name-1250000000
          cos_region: ap-guangzhou
          folder: your-website-output-folder
```

> 如果你使用了 environment secret 请[查看这里](#配置了environment-secret怎么不生效)

### 配置项

| 参数 | 描述 | 必填 | 默认值 |
| --- | --- | --- | --- |
| `provider` | 云服务提供商 (`aliyun` 或 `tencent`) | 否 | `aliyun` |
| `folder` | 包含网站文件的本地目录 | 是 | - |
| `exclude` | 排除的文件或目录 (glob 模式) | 否 | - |
| `incremental` | 是否启用增量更新 | 否 | `true` |
| `skipSetting` | 是否跳过静态页面配置 (index/404) | 否 | `false` |
| `htmlCacheControl` | HTML 文件的 Cache-Control | 否 | `no-cache` |
| `imageCacheControl` | 图片文件的 Cache-Control | 否 | `max-age=864000` |
| `otherCacheControl` | 其他文件的 Cache-Control | 否 | `max-age=2592000` |

#### 阿里云特有参数 (provider: aliyun)

- `accessKeyId`: **必填**
- `accessKeySecret`: **必填**
- `endpoint`: **必填**, 支持指定 protocol, 例如`https://example.org`或者`http://example.org`
- `bucket`: **必填**,部署网站的 bucket, 用于存放网站的资源
- `cname`: 默认`false`. 若`endpoint`填写自定义域名/bucket 域名，需设置为`true`. (使用 CDN 的场景下，不推荐使用自定义域名)

#### 腾讯云特有参数 (provider: tencent)

- `cos_secret_id`: **必填**
- `cos_secret_key`: **必填**
- `cos_bucket`: **必填**, 存储桶名称 (如 `example-1250000000`)
- `cos_region`: **必填**, 存储桶地域 (如 `ap-guangzhou`)

## incremental

**开启`incremental`**
上传文件到 OSS 后，还会将文件的`ContentMD5`和`Cache-Control`收集到名为`.actioninfo`的私有文件中。当再次触发 action 的时候，会将待上传的文件信息与`.actioninfo`中记录的信息比对，信息未发生变化的文件将跳过上传步骤，只进行增量上传。且在上传之后，根据`.actioninfo`和已上传的文件信息，将 OSS 中多余的文件进行删除。

> `.actioninfo` 记录了上一次 action 执行时，所上传的文件信息。私有，不可公共读写。

**关闭`incremental`** 或 OSS 中不存在`.actioninfo`文件

会执行如下步骤

1. 清除所有 OSS 中已有的文件
2. 上传新的文件到 OSS 中

> **计划未来优化这个步骤，优化后，先上传新的文件到 OSS 中，再 diff 删除多余的文件。**

## Cache-Control

为上传的资源默认设置的`Cache-Control`如下

|资源类型 | Cache-Control|
|----| ----|
|.html|no-cache|
|.png/jpg...(图片资源)|max-age=864000(10days)|
|other|max-age=2592000(30days)|

## 静态页面配置

默认的，action 会将阿里云 OSS 的静态页面配置成如下
![2020-08-06-03-18-25](https://image.fangbinwei.cn/github/aliyun-oss-website-action/2020-08-06-03-18-25_05d556d8.png)

若不需要 action 来设置，可以配置`skipSetting`为`true`

## exclude

如果`folder`下的某些文件不需要上传

```yml
    - name: exclude some files
      uses: eallion/s3-deploy-action@v1
      with:
        folder: dist
        exclude: |
          tmp.txt
          tmp/
          tmp2/*.txt
          tmp2/*/*.txt
      # match dist/tmp.txt
      # match dist/tmp/
      # match dist/tmp2/a.txt
      # match dist/tmp2/a/b.txt, not match dist/tmp2/tmp3/a/b.txt
```

> 不支持`**`

或者

```yml
- name: Clean files before upload
  run: rm -f dist/tmp.txt
```

## Docker image / Composite Action

本项目现已支持 Composite Action 模式，无需 Docker 镜像即可快速运行。

如果需要使用 Docker 镜像：

```yml
    - name: upload files to OSS
      uses: docker://eallion/s3-deploy-action:v1
      # 使用 env 而不是 with, 参数可以见本项目的 action.yml
      env:
          ACCESS_KEY_ID: ${{ secrets.ACCESS_KEY_ID }}
          ACCESS_KEY_SECRET: ${{ secrets.ACCESS_KEY_SECRET }}
          BUCKET: your-bucket-name
          ENDPOINT: ali-oss-endpoint
          FOLDER: your-website-output-folder
```

## Demo

### 部署 VuePress 项目

```yml

name: deploy vuepress

on:
  push:
    branches:
      - master

jobs:
  build:

    runs-on: ubuntu-latest
    steps:
      # load repo to /github/workspace
    - uses: actions/checkout@v2
      with:
          repository: eallion/blog
          fetch-depth: 0
    - name: Use Node.js
      uses: actions/setup-node@v1
      with:
        node-version: '12'
    - run: npm install yarn@1.22.4 -g
    - run: yarn install
    # 打包文档命令
    - run: yarn docs:build
    - name: upload files to OSS
      uses: eallion/s3-deploy-action@v1
      with:
          accessKeyId: ${{ secrets.ACCESS_KEY_ID }}
          accessKeySecret: ${{ secrets.ACCESS_KEY_SECRET }}
          bucket: "your-bucket-name"
          endpoint: "oss-cn-shanghai.aliyuncs.com" 
          folder: ".vuepress/dist"
```

具体可以参考本项目的[workflow](.github/workflows/test.yml), npm/yarn配合`action/cache`加速依赖安装

### Vue

[see here](https://github.com/eallion/oss-website-demo-spa-vue)

```yml
- name: upload files to OSS
      uses: eallion/s3-deploy-action@v1
      with:
          accessKeyId: ${{ secrets.ACCESS_KEY_ID }}
          accessKeySecret: ${{ secrets.ACCESS_KEY_SECRET }}
          bucket: website-spa-vue-demo
          endpoint: oss-spa-demo.fangbinwei.cn
          cname: true
          folder: dist
          notFoundPage: index.html
          htmlCacheControl: no-cache
          imageCacheControl: max-age=864001
          otherCacheControl: max-age=2592001
```

## FAQ

### 配合 CDN 使用时，OSS 更新后，CDN 未刷新

开启 OSS 提供的 CDN 缓存自动刷新功能，将触发操作配置为`PutObject`, `DeleteObject`.

![2020-12-13-23-51-28](https://image.fangbinwei.cn/github/aliyun-oss-website-action/2020-12-13-23-51-28_2c310155.png)

![2020-12-13-23-51-55](https://image.fangbinwei.cn/github/aliyun-oss-website-action/2020-12-13-23-51-55_5fe79a54.png)

### `endpoint`使用自定义域名，但是无法上传

1. 如果`endpoint`的域名 CNAME 记录为阿里云 CDN, CDN 是否配置了 http 强制跳转 https? 若配置了，需要在`endpoint`中指定 https, 即`endpoint`为`https://example.org`

2. 如果`endpoint`的域名 CNAME 记录为阿里云 CDN, 在 CDN 为加速范围为全球时有遇到过如下报错`The bucket you are attempting to access must be addressed using the specified endpoint. Please send all future requests to this endpoint.`, 则`endpoint`不能使用自定义域名，使用 OSS 源站的 endpoint.

### 配置了 environment secret 怎么不生效

![2021-05-21-16-47-59](https://image.fangbinwei.cn/github/aliyun-oss-website-action/2021-05-21-16-47-59_affec2b0.png)

如果使用 environment secret, 那么需要如下类似的配置

```diff

jobs:
  build:
    runs-on: ubuntu-latest
+    environment: your-environment-name

```

## Local Testing (本地测试)

如果您已安装 Go 环境，可以在本地直接运行进行测试。

1. **准备测试数据**：

    ```bash
    mkdir -p public
    echo "Hello World" > public/index.html
    ```

2. **设置环境变量** (以腾讯云为例)：

    ```bash
    export FOLDER=./public
    export PROVIDER=tencent
    export COS_SECRET_ID=您的 SecretId
    export COS_SECRET_KEY=您的 SecretKey
    export COS_BUCKET=您的 BucketName
    export COS_REGION=您的 Region
    export INCREMENTAL=true
    ```

3. **运行**：

    ```bash
    go run main.go
    ```
