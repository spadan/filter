- 接口

1. Loader 数据加载接口，业务自定义实现
2. Filter 过滤接口，业务自定义实现
3. Manager 管理器，负责管理所有的Loader和Filter，并按需编排执行

- Manager实现  
核心逻辑  
内部将Loader和Filter封装成统一的Node，Node通过持有父子节点在逻辑上形成一个完整的DAG