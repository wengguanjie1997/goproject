# TG机器人查询AWS-Lightsail流量



## 使用方法

```bash
git clone https://github.com/wengguanjie1997/goproject.git
cd telegram-aws
# 配置config.yaml里面的配置信息
vim config.yaml
# 制作docker镜像
docker build -t tgaws:v1 .
# 运行docker
docker run -itd --name tgcheck tgaws:v1 
# 或者挂载配置文件运行
docker run -itd -v ./config.yaml:/config.yaml --name tgcheck tgaws:v1

```

