# MergeTree原理解析

表引擎是ClickHouse设计实现中的一大特色。可以说，是表引擎决定了一张数据表最终的“性格”，比如数据表拥有何种特性、数据以何种形式被存储以及如何被加载。ClickHouse拥有非常庞大的表引擎体系，截至本书完成时，其共拥有合并树、外部存储、内存、文件、接口和其他6大类20多种表引擎。而在这众多的表引擎中，又属合并树（MergeTree）表引擎及其家族系列（*MergeTree）最为强大，在生产环境的绝大部分场景中，都会使用此系列的表引擎。因为只有合并树系列的表引擎才支持主键索引、数据分区、数据副本和数据采样这些特性，同时也只有此系列的表引擎支持ALTER相关操作。

合并树家族自身也拥有多种表引擎的变种。其中MergeTree作为家族中最基础的表引擎，提供了主键索引、数据分区、数据副本和数据采样等基本能力，而家族中其他的表引擎则在MergeTree的基础之上各有所长。例如ReplacingMergeTree表引擎具有删除重复数据的特性，而SummingMergeTree表引擎则会按照排序键自动聚合数据。如果给合并树系列的表引擎加上Replicated前缀，又会得到一组支持数据副本的表引擎，例如ReplicatedMergeTree、ReplicatedReplacingMergeTree、ReplicatedSummingMergeTree等。

合并树表引擎家族如图6-1所示。

![合并树表引擎家族](images/%E5%90%88%E5%B9%B6%E6%A0%91%E8%A1%A8%E5%BC%95%E6%93%8E%E5%AE%B6%E6%97%8F-16540482760811.png)

虽然合并树的变种很多，但MergeTree表引擎才是根基。作为合并树家族系列中最基础的表引擎，MergeTree具备了该系列其他表引擎共有的基本特征，所以吃透了MergeTree表引擎的原理，就能够掌握该系列引擎的精髓。本章就针对MergeTree的一些基本原理进行解读。

## MergeTree的创建方式与存储结构

**MergeTree在写入一批数据时，数据总会以数据片段的形式写入磁盘，且数据片段不可修改。为了避免片段过多，ClickHouse会通过后台线程，定期合并这些数据片段，属于相同分区的数据片段会被合成一个新的片段。**

这种数据片段往复合并的特点，也正是合并树名称的由来。

## MergeTree的创建方式

创建MergeTree数据表的方法，与我们第4章介绍的定义数据表的方法大致相同，但需要将ENGINE参数声明为MergeTree()，其完整的语法如下所示：

```sql
CREATE TABLE [IF NOT EXISTS] [db_name.]table_name (
name1 [type] [DEFAULT|MATERIALIZED|ALIAS expr],
name2 [type] [DEFAULT|MATERIALIZED|ALIAS expr],
省略...
) ENGINE = MergeTree()
[PARTITION BY expr]
[ORDER BY expr]
[PRIMARY KEY expr]
[SAMPLE BY expr]
[SETTINGS name=value, 省略...]
```

MergeTree表引擎除了常规参数之外，还拥有一些独有的配置选项。

接下来会着重介绍其中几个重要的参数，包括它们的使用方法和工作原理。但是在此之前，还是先介绍一遍它们的作用。

1. PARTITION BY [选填]：分区键，用于指定表数据以何种标准进行分区。分区键既可以是单个列字段，也可以通过元组的形式使用多个列字段，同时它也支持使用列表达式。如果不声明分区键，则ClickHouse会生成一个名为all的分区。合理使用数据分区，可以有效减少查询时数据文件的扫描范围，更多关于数据分区的细节会在6.2节介绍。

   

2. ORDER BY [必填]：排序键，用于指定在一个数据片段内，数据以何种标准排序。默认情况下主键（PRIMARY KEY）与排序键相同。排序键既可以是单个列字段，例如ORDER BY CounterID，也可以通过元组的形式使用多个列字段，例如ORDERBY（CounterID,EventDate）。当使用多个列字段排序时，以ORDERBY（CounterID,EventDate）为例，在单个数据片段内，数据首先会以CounterID排序，相同CounterID的数据再EventDate排序。

   

3. PRIMARY KEY [选填]：主键，顾名思义，声明后会依照主键字段生成一级索引，用于加速表查询。默认情下，主键与排序键(ORDER BY)相同，所以通常直接使用ORDER BY代为指定主键，无须刻意通过PRIMARY KEY声明。所以在一般情况下，在单个数据片段内，数据与一级索引以相同的规则升序排列。与其他数据库不同，MergeTree主键允许存在重复数据（ReplacingMergeTree可以去重）。

   

4. SAMPLE BY [选填]：抽样表达式，用于声明数据以何种标准进行采样。如果使用了此配置项，那么在主键的配置中也需要声明同样的表达式，例如：

   ```sql
   省略...
   ) ENGINE = MergeTree()
   ORDER BY (CounterID, EventDate, intHash32(UserID)
   SAMPLE BY intHash32(UserID)
   ```

   抽样表达式需要配合SAMPLE子查询使用，这项功能对于选取抽样数据十分有用，更多关于抽样查询的使用方法会在第9章介绍。

   

5. SETTINGS：index_granularity [选填]：

   index_granularity对于MergeTree而言是一项非常重要的参数，它表示索引的粒度，默认值为8192。也就是说，MergeTree的索引在默认情况下，每间隔8192行数据才生成一条索引，其具体声明方式如下所示：

   ```sql
   省略...
   ) ENGINE = MergeTree()
   省略...
   SETTINGS index_granularity = 8192;
   ```

   8192是一个神奇的数字，在ClickHouse中大量数值参数都有它的影子，可以被其整除（例如最小压缩块大小min_compress_block_size:65536）。通常情况下并不需要修改此参数，但理解它的工作原理有助于我们更好地使用MergeTree。关于索引详细的工作原理会在后续阐述。

   

6. SETTINGS：index_granularity_bytes [选填]：在19.11版本之前，ClickHouse只支持固定大小的索引间隔，由index_granularity控制，默认为8192。在新版本中，它增加了自适应间隔大小的特性，即根据每一批次写入数据的体量大小，动态划分间隔大小。而数据的体量大小，正是由index_granularity_bytes参数控制的，默认为10M(10×1024×1024)，设置为0表示不启动自适应功能。

   

7. SETTINGS：enable_mixed_granularity_parts [选填]：设置是否开启自适应索引间隔的功能，默认开启。

   

8. SETTINGS：merge_with_ttl_timeout [选填]：从19.6版本开始，MergeTree提供了数据TTL的功能，关于这部分的详细介绍，将留到第7章介绍。

   

9. SETTINGS：storage_policy [选填]：从19.15版本开始，MergeTree提供了多路径的存储策略，关于这部分的详细介绍，同样留到第7章介绍。

## MergeTree的存储结构

MergeTree表引擎中的数据是拥有物理存储的，数据会按照分区目录的形式保存到磁盘之上，其完整的存储结构如图6-2所示。

![MergeTree在磁盘上的物理存储结构](images/MergeTree%E5%9C%A8%E7%A3%81%E7%9B%98%E4%B8%8A%E7%9A%84%E7%89%A9%E7%90%86%E5%AD%98%E5%82%A8%E7%BB%93%E6%9E%84.png)

从图6-2中可以看出，一张数据表的完整物理结构分为3个层级，依次是数据表目录、分区目录及各分区下具体的数据文件。接下来就逐一介绍它们的作用。

