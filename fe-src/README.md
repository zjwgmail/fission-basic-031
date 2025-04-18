# 淘宝镜像源

切换不同镜像源
npm config set registry https://registry.npmmirror.com

检查是否设置成功
npm config get registry

临时使用
npm install <package-name> --registry=https://registry.npmmirror.com

# npm镜像源

切换不同镜像源
npm config set registry https://registry.npmjs.org

检查是否设置成功
npm config get registry

临时使用
npm install <package-name> --registry=https://registry.npmjs.org

```
1、邀请好友链接  把html输出给我，以及配置好，根据用户信息接口查询必要的参数，waName
后端动态的url
https://game.laotielaila.com/events/goldenmonthvn/index?code=02&gpt=8 
我会根据code下发不同的页面可能是invitationlang02也可能是invitationlang04
等价于下面
https://play.moba5v5.com/events/moba2025wa/invite/invitationlang02?code=a0202djg76&gpt=8

2、奖励链接  
红包1、3、6、9、12、15
3、6、9、12、15 提前生成5个
https://game.laotielaila.com/events/mlbb25031/promotion/?code=sdj1213&gpt=1&lp=1&mode=3


拿到加密code，调用接口查询用户信息
url： /events/mlbb25031gateway/activity/cdk 
参数 {
    param:"加密字符串"
}
用下面的json进行加密成param的值
{
    "code":"sdj1213" //用户id
    "mode":1,//返回cdk,如果不传则返回普通用户信息
}
响应
{
    "waName":"昵称"
    "rallyCode":"sdj1213", //用户唯一code
    "language": "01", //语言
    "channel":"a",  //渠道
    "generation" //代次
    "cdk" :"" //对应mode的cdk
}
```