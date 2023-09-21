# cock 程序唤起/任务调度

一个兼备程序唤起和任务调度程序


## cock （包含 http-dashboard）

自行编译，需要 `go1.21` `node`
按照以下方式编译获取可执行文件`cock`。
```shell
git clone  https://github.com/leancodebox/cock.git 
cd cock 
cd actor 
npm i
npm run build
cd ..
go install 
```

## cock-cli

如果你有 `go1.21` 以上的环境，你可以尝试使用下面命令快速开始。

```shell
go install github.com/leancodebox/cock-cli@latest 
```

执行 `cock-cli` 后会判断当前目录是否存在 `jobConfig.json`，如果没有会提示是否生成默认配置，无论是否生成默认配置，本次都不会真正去执行程序唤起/任务调度。  

可以在生成后修改完毕配置，再次执行 `cock-cli` 运行任务调度。相关参数配置如下。

## 参数说明


|                 key                  |  value   |                    desc                    |
| :----------------------------------: | :------: |:------------------------------------------:|
|               `config`               | `object` |                    基础配置                    |
|          `config.dashboard`          | `object` |                    面板配置                    |
|       `config.dashboard.port`        |  `int`   | 端口，小于1，不开启（cock-cli无论配置与否，都不会开启dashboardÏ） |
|            `residentTask`            | `array`  |                    常驻任务                    |
|       `residentTask.[]jobName`       | `string` |                    任务名                     |
|       `residentTask.[]binPath`       | `string` |           可执行文件路径，或者环境变量中的可执行命令            |
|       `residentTask.[]params`        | `array`  |                     参数                     |
|      `residentTask.[]params.[]`      | `string` |                    参数列表                    |
|         `residentTask.[]dir`         | `string` |                    执行目录                    |
|         `residentTask.[]run`         |  `bool`  |    是否开启 ，true 开启 false 不开启，可以在web中开启关闭     |
|       `residentTask.[]options`       | `object` |                     选项                     |
| `residentTask.[]options.outputType`  |  `int`   |             输出模式 0 标准输出 1 文件输出             |
| `residentTask.[]options.outputPath`  | `string` |                    输出路径                    |
|           `scheduledTask`            | `array`  |                    任务名                     |
|      `scheduledTask.[]jobName`       | `string` |           可执行文件路径，或者环境变量中的可执行命令            |
|      `scheduledTask.[]binPath`       | `string` |                     参数                     |
|       `scheduledTask.[]params`       | `array`  |                    参数列表                    |
|     `scheduledTask.[]params.[]`      | `string` |                    执行目录                    |
|        `scheduledTask.[]spec`        | `string` |        crontab 格式周期 例如 `* * * * * `        |
|        `scheduledTask.[]run`         |  `bool`  |    是否开启 ，true 开启 false 不开启，可以在web中开启关闭     |
|       `scheduledTask.[]options`       | `object` |                     选项                     |
| `scheduledTask.[]options.outputType` |  `int`   |             输出模式 0 标准输出 1 文件输出             |
| `scheduledTask.[]options.outputPath` | `string` |                    输出路径                    |

