
## 解码 CryptoJS
- 一般使用了```CryptoJS```的，该关键字不会被```混肴```
- 如果没混肴
- 如果```JS```可以格式化，直接格式化
- 然后```console.log```
- 如果```console.log``` 没了的
- 直接在文件最上方定义``` const fuckLog = console.log; ```
- 然后在```CryptoJS```上下文打日志``` fuckLog(xxx) ```
- 如果能打印最终结果那就找对地方了
- 如果找对了，就把该方法转为可读的
- 然后直接转译到其他语言，或者使用相关```JS```库桥接调用
- 如果上述没找到，那么恭喜
- 老费事了
- 以上都不行就看控制台```network```有没有请求视频文件执行相关的```JS```上下文位置
- 如果有，使用上述逻辑再试
- 如果没有
- 可能需要挨个位置打断点找了
- 如果不能格式化```JS```代码的
- 理论上只是不能格式化到特定位置的代码，部分代码还是可以单独拎出来格式化的
- 有```debugger```的阻止调试的
- 查询```debugger```位置并删除