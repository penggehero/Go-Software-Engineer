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

1. partition：分区目录，余下各类数据文件（primary.idx、[Column].mrk、[Column].bin等）都是以分区目录的形式被组织存放的，属于相同分区的数据，最终会被合并到同一个分区目录，而不同分区的数据，永远不会被合并在一起。更多关于数据分区的细节会在6.2节阐述。

   

2. checksums.txt：校验文件，使用二进制格式存储。它保存了余下各类文件(primary.idx、count.txt等)的size大小及size的哈希值，用于快速校验文件的完整性和正确性。

   

3. columns.txt：列信息文件，使用明文格式存储。用于保存此数据分区下的列字段信息，例如

   ```sh
   $ cat columns.txt
   columns format version: 1
   4 columns:
   'ID' String
   'URL' String
   'Code' String
   'EventTime' Date
   ```

   

4. count.txt：计数文件，使用明文格式存储。用于记录当前数据分区目录下数据的总行数，例如：

   ```shell
   $ cat count.txt
   8
   ```

   

5. primary.idx：一级索引文件，使用二进制格式存储。用于存放稀疏索引，一张MergeTree表只能声明一次一级索引（通过ORDERBY或者PRIMARY KEY）。借助稀疏索引，在数据查询的时能够排除主键条件范围之外的数据文件，从而有效减少数据扫描范围，加速查询速度。更多关于稀疏索引的细节与工作原理会在6.3节阐述。

   

6. [Column].bin：数据文件，使用压缩格式存储，默认为LZ4压缩格式，用于存储某一列的数据。由于MergeTree采用列式存储，所以每一个列字段都拥有独立的.bin数据文件，并以列字段名称命名（例如CounterID.bin、EventDate.bin等）。更多关于数据存储的细节会在6.5节阐述。

   

7. [Column].mrk：列字段标记文件，使用二进制格式存储。标记文件中保存了.bin文件中数据的偏移量信息。标记文件与稀疏索引对齐，又与.bin文件一一对应，所以MergeTree通过标记文件建立了primary.idx稀疏索引与.bin数据文件之间的映射关系。即首先通过稀疏索引（primary.idx）找到对应数据的偏移量信息（.mrk），再通过偏移量直接从.bin文件中读取数据。由于.mrk标记文件与.bin文件一一对应，所以MergeTree中的每个列字段都会拥有与其对应的.mrk标记文件（例如CounterID.mrk、EventDate.mrk等）。更多关于数据标记的细节会在6.6节阐述。

   

8. [Column].mrk2：如果使用了自适应大小的索引间隔，则标记文件会以.mrk2命名。它的工作原理和作用与.mrk标记文件相同。

   

9. partition.dat与minmax_[Column].idx：如果使用了分区键，例如PARTITION BY EventTime，则会额外生成partition.dat与minmax索引文件，它们均使用二进制格式存储。partition.dat用于保存当前分区下分区表达式最终生成的值；而minmax索引用于记录当前分区下分区字段对应原始数据的最小和最大值。例如EventTime字段对应的原始数据为2019-05-01、2019-05-05，分区表达式为PARTITIONBY toYYYYMM(EventTime)。partition.dat中保存的值将会是2019-05，而minmax索引中保存的值将会是2019-05-012019-05-05。

   **在这些分区索引的作用下，进行数据查询时能够快速跳过不必要的数据分区目录，从而减少最终需要扫描的数据范围。**

   

10. skp_idx_[Column].idx与skp_idx_[Column].mrk：如果在建表语句中声明了二级索引，则会额外生成相应的二级索引与标记文件，它们同样也使用二进制存储。二级索引在ClickHouse中又称跳数索引，目前拥有minmax、set、ngrambf_v1和tokenbf_v1四种类型。这些索引的最终目标与一级稀疏索引相同，都是为了进一步减少所需扫描的数据范围，以加速整个查询过程。更多关于二级索引的细节会在6.4节阐述。

## 数据分区

通过先前的介绍已经知晓在MergeTree中，数据是以分区目录的形式进行组织的，每个分区独立分开存储。借助这种形式，在对MergeTree进行数据查询时，可以有效跳过无用的数据文件，只使用最小的分区目录子集。这里有一点需要明确，在ClickHouse中，数据分区（partition）和数据分片（shard）是完全不同的概念。**数据分区是针对本地数据而言的，是对数据的一种纵向切分。**MergeTree并不能依靠分区的特性，将一张表的数据分布到多个ClickHouse服务节点。而**横向切分是数据分片（shard）的能力**，关于这一点将在后续章节介绍。本节将针对“数据分区目录具体是如何运作的”这一问题进行分析

### 数据的分区规则

MergeTree数据分区的规则由分区ID决定，而具体到每个数据分区所对应的ID，则是由分区键的取值决定的。分区键支持使用任何一个或一组字段表达式声明，其业务语义可以是年、月、日或者组织单位等任何一种规则。针对取值数据类型的不同，分区ID的生成逻辑目前拥有四种规则：

1. 不指定分区键：如果不使用分区键，即不使用PARTITION BY声明任何分区表达式，则分区ID默认取名为all，所有的数据都会被写入这个all分区。

   

2. 使用整型：如果分区键取值属于整型（兼容UInt64，包括有符号整型和无符号整型），且无法转换为日期类型YYYYMMDD格式，则直接按照该整型的字符形式输出，作为分区ID的取值

   

3. 使用日期类型：如果分区键取值属于日期类型，或者是能够转换为YYYYMMDD格式的整型，则使用按照YYYYMMDD进行格式化后的字符形式输出，并作为分区ID的取值

   

4. 使用其他类型：如果分区键取值既不属于整型，也不属于日期类型，例如String、Float等，则通过128位Hash算法取其Hash值作为分区ID的取值。数据在写入时，会对照分区ID落入相应的数据分区，表6-1列举了分区ID在不同规则下的一些示例。

![ID在不同规则下的示例](images/ID%E5%9C%A8%E4%B8%8D%E5%90%8C%E8%A7%84%E5%88%99%E4%B8%8B%E7%9A%84%E7%A4%BA%E4%BE%8B.png)

如果通过元组的方式使用多个分区字段，则分区ID依旧是根据上述规则生成的，只是多个ID之间通过“-”符号依次拼接。例如按照上述表格中的例子，使用两个字段分区：

```
PARTITION BY (length(Code),EventTime)
```

则最终的分区ID会是下面的模样：

```
2-20190501
2-20190611
```



### 分区目录的命名规则

通过上一小节的介绍，我们已经知道了分区ID的生成规则。但是如果进入数据表所在的磁盘目录后，会发现MergeTree分区目录的完整物理名称并不是只有ID而已，在ID之后还跟着一串奇怪的数字，例如201905110。那么这些数字又代表着什么呢？

众所周知，对于MergeTree而言，它最核心的特点是其分区目录的合并动作。但是我们可曾想过，从分区目录的命名中便能够解读出它的合并逻辑。在这一小节，我们会着重对命名公式中各分项进行解读，而关于具体的目录合并过程将会留在后面小节讲解。一个完整分区目录的命名公式如下所示：

```
PartitionIDMinBlockNumMaxBlockNumLevel
```

如果对照着示例数据，那么数据与公式的对照关系会如同图6-3所示一般。

![命名公式与样例数据的对照关系](images/%E5%91%BD%E5%90%8D%E5%85%AC%E5%BC%8F%E4%B8%8E%E6%A0%B7%E4%BE%8B%E6%95%B0%E6%8D%AE%E7%9A%84%E5%AF%B9%E7%85%A7%E5%85%B3%E7%B3%BB.png)

上图中，201905表示分区目录的ID；1_1分别表示最小的数据块编号与最大的数据块编号；而最后的_0则表示目前合并的层级。接下来开始分别解释它们的含义：

1. PartitionID：分区ID，无须多说，关于分区ID的规则在上一小节中已经做过详细阐述了。

   

2. MinBlockNum和MaxBlockNum：顾名思义，最小数据块编号与最大数据块编号。ClickHouse在这里的命名似乎有些歧义，很容易让人与稍后会介绍到的数据压缩块混淆。但是本质上它们毫无关系，这里的BlockNum是一个整型的自增长编号。如果将其设为n的话，那么计数n在单张MergeTree数据表内全局累加，n从1开始，每当新创建一个分区目录时，计数n就会累积加1。对于一个新的分区目录而言，MinBlockNum与MaxBlockNum取值一样，同等于n，例如201905_1_1_0、201906_2_2_0以此类推。但是也有例外，当分区目录发生合并时，对于新产生的合并目录MinBlockNum与MaxBlockNum有着另外的取值规则。对于合并规则，我们留到下一小节再详细讲解。

   

3. Level：合并的层级，可以理解为某个分区被合并过的次数，或者这个分区的年龄。数值越高表示年龄越大。Level计数BlockNum有所不同，它并不是全局累加的。对于每一个新创建的分区目录而言，其初始值均为0。之后，以分区为单位，如果相同分区发生合并动作，则在相应分区内计数累积加1。

### 分区目录的合并过程

MergeTree的分区目录和传统意义上其他数据库有所不同。首先，MergeTree的分区目录并不是在数据表被创建之后就存在的，而是在数据写入过程中被创建的。也就是说如果一张数据表没有任何数据，那么也不会有任何分区目录存在。其次，它的分区目录在建立之后也并不是一成不变的。在其他某些数据库的设计中，追加数据后目录自身不会发生变化，只是在相同分区目录中追加新的数据文件。而MergeTree完全不同，伴随着每一批数据的写入（一次INSERT语句），MergeTree都会生成一批新的分区目录。即便不同批次写入的数据属于相同分区，也会生成不同的分区目录。也就是说，对于同一个分区而言，也会存在多个分区目录的情况。在之后的某个时刻（写入后的10～15分钟，也可以手动执行optimize查询语句），ClickHouse会通过后台任务再将属于相同分区的多个目录合并成一个新的目录。已经存在的旧分区目录并不会立即被删除，而是在之后的某个时刻通过后台任务被删除（默认8分钟）。

属于同一个分区的多个目录，在合并之后会生成一个全新的目录，目录中的索引和数据文件也会相应地进行合并。新目录名称的合并方式遵循以下规则，其中：

- MinBlockNum：取同一分区内所有目录中最小的MinBlockNum值。
- MaxBlockNum：取同一分区内所有目录中最大的MaxBlockNum值。
- Level：取同一分区内最大Level值并加1。

合并目录名称的变化过程如图6-4所示。

![名称变化过程](images/%E5%90%8D%E7%A7%B0%E5%8F%98%E5%8C%96%E8%BF%87%E7%A8%8B.png)

在图6-4中，partition_v5测试表按日期字段格式分区，即PARTITION BY toYYYYMM（EventTime），T表示时间。假设在T0时刻，首先分3批（3次INSERT语句）写入3条数据人：

```sql
INSERT INTO partition_v5 VALUES (A, c1, '2019-05-01')
INSERT INTO partition_v5 VALUES (B, c1, '2019-05-02')
INSERT INTO partition_v5 VALUES (C, c1, '2019-06-01')
```

按照目录规，上述代码会创建3个分区目录。分区目录的名称由PartitionID、MinBlockNum、MaxBlockNum和Level组成，其中PartitionID根据6.2.1节介绍的生成规则，3个分区目录的ID依次为201905、201905和201906。而对于每个新建的分区目录而言，它们的MinBlockNum与MaxBlockNum取值相同，均来源于表内全局自增的BlockNum。BlockNum初始为1，每次新建目录后累计加1。所以，3个分区目录的MinBlockNum与MaxBlockNum依次为0_0、1_1和2_2。最后是Level层级，每个新建的分区目录初始Level都是0。所以3个分区目录的最终名称分别是201905_1_1_0、201905_2_2_0和201906_3_3_0。

假设在T1时刻，MergeTree的合并动作开始了，那么属于同一分区的201905_1_1_0与201905_2_2_0目录将发生合并。从图6-4所示过程中可以发现，合并动作完成后，生成了一个新的分区201905_1_2_1。根据本节所述的合并规则，其中，MinBlockNum取同一分区内所有目录中最小的MinBlockNum值，所以是1；MaxBlockNum取同一分区内所有目录中最大的MaxBlockNum值，所以是2；而Level则取同一分区内，最大Level值加1，所以是1。而后续T2时刻的合并规则，只是在重复刚才所述的过程而已。

至此，大家已经知道了分区ID、目录命名和目录合并的相关规则。最后，再用一张完整的示例图作为总结，描述MergeTree分区目录从创建、合并到删除的整个过程，如图6-5所示。

![分区目录创建、合并、删除的过程](images/%E5%88%86%E5%8C%BA%E7%9B%AE%E5%BD%95%E5%88%9B%E5%BB%BA%E3%80%81%E5%90%88%E5%B9%B6%E3%80%81%E5%88%A0%E9%99%A4%E7%9A%84%E8%BF%87%E7%A8%8B.png)

从图6-5中应当能够发现，分区目录在发生合并之后，旧的分区目录并没有被立即删除，而是会存留一段时间。但是旧的分区目录已不再是激活状态（active=0），所以在数据查询时，它们会被自动过滤掉。

## 一级索引

MergeTree的主键使用PRIMARY KEY定义，待主键定义之后，MergeTree会依据index_granularity间隔（默认8192行），为数据表生成一级索引并保存至primary.idx文件内，索引数据按照PRIMARYKEY排序。相比使用PRIMARY KEY定义，更为常见的简化形式是通过ORDER BY指代主键。在此种情形下，PRIMARY KEY与ORDER BY定义相同，所以索引（primary.idx）和数据（.bin）会按照完全相同的规则排序。对于PRIMARY KEY与ORDER BY定义有差异的应用场景在SummingMergeTree引擎章节部分会所有介绍，而关于数据文件的更多细节，则留在稍后的6.5节介绍，本节重点讲解一级索引部分。

### 稀疏索引

primary.idx文件内的一级索引采用稀疏索引实现。此时有人可能会问，既然提到了稀疏索引，那么是不是也有稠密索引呢？还真有！稀疏索引和稠密索引的区别如图6-6所示。

![稀疏索引与稠密索引的区别](images/%E7%A8%80%E7%96%8F%E7%B4%A2%E5%BC%95%E4%B8%8E%E7%A8%A0%E5%AF%86%E7%B4%A2%E5%BC%95%E7%9A%84%E5%8C%BA%E5%88%AB.png)

简单来说，在稠密索引中每一行索引标记都会对应到一行具体的数据记录。而在稀疏索引中，每一行索引标记对应的是一段数据，而不是一行。用一个形象的例子来说明：如果把MergeTree比作一本书，那么稀疏索引就好比是这本书的一级章节目录。一级章节目录不会具体对应到每个字的位置，只会记录每个章节的起始页码。

稀疏索引的优势是显而易见的，它仅需使用少量的索引标记就能够记录大量数据的区间位置信息，且数据量越大优势越为明显。以默认的索引粒度（8192）为例，MergeTree只需要12208行索引标记就能为1亿行数据记录提供索引。由于稀疏索引占用空间小，所以primary.idx内的索引数据常驻内存，取用速度自然极快。

### 索引粒度

在先前的篇幅中已经数次出现过index_granularity这个参数了，它表示索引的粒度。虽然在新版本中，ClickHouse提供了自适应粒度大小的特性，但是为了便于理解，仍然会使用固定的索引粒度（默认8192）进行讲解。索引粒度对MergeTree而言是一个非常重要的概念，因此很有必要对它做一番深入解读。索引粒度就如同标尺一般，会丈量整个数据的长度，并依照刻度对数据进行标注，最终将数据标记成多个间隔的小段，如图6-7所示。

![MergeTree按照索引粒度](images/MergeTree%E6%8C%89%E7%85%A7%E7%B4%A2%E5%BC%95%E7%B2%92%E5%BA%A6.png)

数据以index_granularity的粒度（默认8192）被标记成多个小的区间，其中每个区间最多8192行数据。MergeTree使用MarkRange表示一个具体的区间，并通过start和end表示其具体的范围。index_granularity的命名虽然取了索引二字，但它不单只作用于一级索引（.idx），同时也会影响数据标记（.mrk）和数据文件（.bin）。因为仅有一级索引自身是无法完成查询工作的，它需要借助数据标记才能定位数据，所以一级索引和数据标记的间隔粒度相同（同为index_granularity行），彼此对齐。而数据文件也会依照index_granularity的间隔粒度生成压缩数据块。关于数据文件和数据标记的细节会在后面说明。

### 索引数据的生成规则

由于是稀疏索引，所以MergeTree需要间隔index_granularity行数据才会生成一条索引记录，其索引值会依据声明的主键字段获取。图6-8所示是对照测试表hits_v1中的真实数据具象化后的效果。hits_v1使用年月分区（PARTITION BY toYYYYMM(EventDate)），所以2014年3月份的数据最终会被划分到同一个分区目录内。如果使用CounterID作为主键（ORDER BY CounterID），则每间隔8192行数据就会取一次CounterID的值作为索引值，索引数据最终会被写入primary.idx文件进行保存。

![测试表hits_v1具象化后的效果](images/%E6%B5%8B%E8%AF%95%E8%A1%A8hits_v1%E5%85%B7%E8%B1%A1%E5%8C%96%E5%90%8E%E7%9A%84%E6%95%88%E6%9E%9C.png)

例如第0(8192*0)行CounterID取值57，第8192(8192*1)行CounterID取值1635，而第16384(8192*2)行CounterID取值3266，最终索引数据将会是5716353266。

从图6-8中也能够看出，MergeTree对于稀疏索引的存储是非常紧凑的，索引值前后相连，按照主键字段顺序紧密地排列在一起。不仅此处，ClickHouse中很多数据结构都被设计得非常紧凑，比如其使用位读取替代专门的标志位或状态码，可以不浪费哪怕一个字节的空间。以小见大，这也是ClickHouse为何性能如此出众的深层原因之一。

如果使用多个主键，例如ORDER BY(CounterID,EventDate)，则每间隔8192行可以同时取CounterID与EventDate两列的值作为索引值，具体如图6-9所示。

![使用CounterID和EventDate作为主键](images/%E4%BD%BF%E7%94%A8CounterID%E5%92%8CEventDate%E4%BD%9C%E4%B8%BA%E4%B8%BB%E9%94%AE.png)

### 索引的查询过程

在介绍了上述关于索引的一些概念之后，接下来说明索引具体是如何工作的。首先，我们需要了解什么是MarkRange。MarkRange在ClickHouse中是用于定义标记区间的对象。通过先前的介绍已知，MergeTree按照index_granularity的间隔粒度，将一段完整的数据划分成了多个小的间隔数据段，一个具体的数据段即是一个MarkRange。MarkRange与索引编号对应，使用start和end两个属性表示其区间范围。通过与start及end对应的索引编号的取值，即能够得到它所对应的数值区间。而数值区间表示了此MarkRange包含的数据范围。

如果只是这么干巴巴地介绍，大家可能会觉得比较抽象，下面用一份示例数据来进一步说明。假如现在有一份测试数据，共192行记录。其中，主键ID为String类型，ID的取值从A000开始，后面依次为A001、A002……直至A192为止。MergeTree的索引粒度index_granularity=3，根据索引的生成规则，primary.idx文件内的索引数据会如图6-10所示。

![192行ID索引的物理存储示意](images/192%E8%A1%8CID%E7%B4%A2%E5%BC%95%E7%9A%84%E7%89%A9%E7%90%86%E5%AD%98%E5%82%A8%E7%A4%BA%E6%84%8F.png)

根据索引数据，MergeTree会将此数据片段划分成192/3=64个小的MarkRange，两个相邻MarkRange相距的步长为1。其中，所有MarkRange（整个数据片段）的最大数值区间为[A000,+inf)，其完整的示意如图6-11所示。

![64个MarkRange与其数值区间范围的示意图](images/64%E4%B8%AAMarkRange%E4%B8%8E%E5%85%B6%E6%95%B0%E5%80%BC%E5%8C%BA%E9%97%B4%E8%8C%83%E5%9B%B4%E7%9A%84%E7%A4%BA%E6%84%8F%E5%9B%BE.png)

在引出了数值区间的概念之后，对于索引的查询过程就很好解释了。索引查询其实就是两个数值区间的交集判断。其中，一个区间是由基于主键的查询条件转换而来的条件区间；而另一个区间是刚才所讲述的与MarkRange对应的数值区间。

整个索引查询过程可以大致分为3个步骤。

1. 生成查询条件区间：首先，将查询条件转换为条件区间。即便是单个值的查询条件，也会被转换成区间的形式，例如下面的例子。

   ```sql
   WHERE ID = 'A003'
   ['A003', 'A003']
   WHERE ID > 'A000'
   ('A000', +inf)
   WHERE ID < 'A188'
   (-inf, 'A188')
   WHERE ID LIKE 'A006%'
   ['A006', 'A007')
   ```



2. 递归交集判断：以递归的形式，依次对MarkRange的数值区间与条件区间做交集判断。从最大的区间[A000,+inf)开始：

   - 如果不存在交集，则直接通过剪枝算法优化此整段MarkRange。
   - 如果存在交集，且MarkRange步长大于8(end-start)，则将此区间进一步拆分成8个子区间（由merge_tree_coarse_index_granularity指定，默认值为8），并重复此规则，继续做递归交集判断。
   - 如果存在交集，且MarkRange不可再分解（步长小于8），则记录MarkRange并返回。

3. 合并MarkRange区间：将最终匹配的MarkRange聚在一起，合并它们的范围。

完整逻辑的示意如图6-12所示。

MergeTree通过递归的形式持续向下拆分区间，最终将MarkRange定位到最细的粒度，以帮助在后续读取数据的时候，能够最小化扫描数据的范围。以图6-12所示为例，当查询条件WHERE ID='A003'的时候，最终只需要读取[A000,A003]和[A003,A006]两个区间的数据,它们对应MarkRange(start:0,end:2)范围，而其他无用的区间都被裁剪掉了。因为MarkRange转换的数值区间是闭区间，所以会额外匹配到临近的一个区间。

## 二级索引

除了一级索引之外，MergeTree同样支持二级索引。二级索引又称跳数索引，由数据的聚合信息构建而成。根据索引类型的不同，其聚合信息的内容也不同。跳数索引的目的与一级索引一样，也是帮助查询时减少数据扫描的范围。

跳数索引在默认情况下是关闭的，需要设置allow_experimental_data_skipping_indices（该参数在新版本中已被取消）才能使用：

```
SET allow_experimental_data_skipping_indices = 1
```

![索引查询完整过程的逻辑示意图](images/%E7%B4%A2%E5%BC%95%E6%9F%A5%E8%AF%A2%E5%AE%8C%E6%95%B4%E8%BF%87%E7%A8%8B%E7%9A%84%E9%80%BB%E8%BE%91%E7%A4%BA%E6%84%8F%E5%9B%BE.png)

跳数索引需要在CREATE语句内定义，它支持使用元组和表达式的形式声明，其完整的定义语法如下所示：

```
INDEX indexname expr TYPE indextype(...) GRANULARITY granularity
```

与一级索引一样，如果在建表语句中声明了跳数索引，则会额外生成相应的索引与标记文件（skp_idx_[Column].idxskp_idx_[Column].mrk）。

### granularity与index_granularity的关系

不同的跳数索引之间，除了它们自身独有的参数之外，还都共同拥有granularity参数。初次接触时，很容易将granularity与index_granularity的概念弄混淆。对于跳数索引而言，index_granularity定义了数据的粒度，而granularity定义了聚合信息汇总的粒度。换言之，granularity定义了一行跳数索引能够跳过多少个index_granularity区间的数据。

要解释清楚granularity的作用，就要从跳数索引的数据生成规则说起，其规则大致是这样的：首先，按照index_granularity粒度间隔将数据划分成n段，总共有[0,n-1]个区间（n=total_rows/index_granularity，向上取整）。接着，根据索引定义时声明的表达式，从0区间开始，依次按index_granularity粒度从数据中获取聚合信息，每次向前移动1步(n+1)，聚合信息逐步累加。最后，当移动granularity次区间时，则汇总并生成一行跳数索引数据。

以minmax索引为例，它的聚合信息是在一个index_granularity区间内数据的最小和最大极值。以下图为例，假设index_granularity=8192且granularity=3，则数据会按照index_granularity划分为n等份，MergeTree从第0段分区开始，依次获取聚合信息。当获取到第3个分区时（granularity=3），则汇总并会生成第一行minmax索引（前3段minmax极值汇总后取值为[1,9]），如图6-13所示。

![跳数索引granularity与index_granularity的关系](images/%E8%B7%B3%E6%95%B0%E7%B4%A2%E5%BC%95granularity%E4%B8%8Eindex_granularity%E7%9A%84%E5%85%B3%E7%B3%BB.png)

### 跳数索引的类型

目前，MergeTree共支持4种跳数索引，分别是minmax、set、ngrambf_v1和tokenbf_v1。一张数据表支持同时声明多个跳数索引，例如：

```sql
CREATE TABLE skip_test (
ID String,
URL String,
Code String,
EventTime Date,
INDEX a ID TYPE minmax GRANULARITY 5,
INDEX b（length(ID) * 8） TYPE set(2) GRANULARITY 5,
INDEX c（ID，Code） TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 5,
INDEX d ID TYPE tokenbf_v1(256, 2, 0) GRANULARITY 5
) ENGINE = MergeTree()
省略...
```

接下来，就借助上面的例子逐个介绍这几种跳数索引的用法：

1. minmax：minmax索引记录了一段数据内的最小和最大极值，其索引的作用类似分区目录的minmax索引，能够快速跳过无用的数据区间，示例如下所示：

   ```
   INDEX a ID TYPE minmax GRANULARITY 5
   ```

   上述示例中minmax索引会记录这段数据区间内ID字段的极值。极值的计算涉及每5个indexgranularity区间中的数据。

2. set：set索引直接记录了声明字段或表达式的取值（唯一值，无重复），其完整形式为set(max_rows)，其中max_rows是一个阈值，表示在一个index_granularity内，索引最多记录的数据行数。如果max_rows=0，则表示无限制，例如：

   ```
   INDEX b（length(ID) * 8） TYPE set(100) GRANULARITY 5
   ```

   上述示例中set索引会记录数据中ID的长度*8后的取值。其中，每个index_granularity内最多记录100条。

3. ngrambf_v1：ngrambf_v1索引记录的是数据短语的布隆表过滤器，只支持String和FixedString数据类型。ngrambf_v1只能够提升in、notIn、like、equals和notEquals查询的性能，其完整形式为ngrambf_v1(n,size_of_bloom_filter_in_bytes,number_of_hash_functions,random_seed)。这些参数是一个布隆过滤器的标准输入，如果你接触过布隆过滤器，应该会对此十分熟悉。它们具体的含义如下：

   - n：token长度，依据n的长度将数据切割为token短语
   - size_of_bloom_filter_in_bytes：布隆过滤器的大小
   - number_of_hash_functions：布隆过滤器中使用Hash函数的个数
   - random_seed：Hash函数的随机种子。

   例如在下面的例子中，ngrambf_v1索引会依照3的粒度将数据切割成短语token，token会经过2个Hash函数映射后再被写入，布隆过滤器大小为256字节。

   ```
   INDEX c（ID，Code） TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 5
   ```

4. tokenbf_v1：tokenbf_v1索引是ngrambf_v1的变种，同样也是一种布隆过滤器索引。tokenbf_v1除了短语token的处理方法外，其他与ngrambf_v1是完全一样的。tokenbf_v1会自动按照非字符的、数字的字符串分割token，具体用法如下所示：

   ```
   INDEX d ID TYPE tokenbfv1(25620) GRANULARITY 5
   ```

## 数据存储

此前已经多次提过，在MergeTree中数据是按列存储的。但是前面的介绍都较为抽象，具体到存储的细节、MergeTree是如何工作的，读者心中难免会有疑问。数据存储，就好比一本书中的文字，在排版时，绝不会密密麻麻地把文字堆满，这样会导致难以阅读。更为优雅的做法是，将文字按段落的形式精心组织，使其错落有致。本节将进一步介绍MergeTree在数据存储方面的细节，尤其是其中关于压缩数据块的概念。

### 各列独立存储

在MergeTree中，数据按列存储。而具体到每个列字段，数据也是独立存储的，每个列字段都拥有一个与之对应的.bin数据文件。也正是这些.bin文件，最终承载着数据的物理存储。数据文件以分区目录的形式被组织存放，所以在.bin文件中只会保存当前分区片段内的这一部分数据，其具体组织形式已经在图6-2中展示过。按列独立存储的设计优势显而易见：一是可以更好地进行数据压缩（相同类型的数据放在一起，对压缩更加友好），二是能够最小化数据扫描的范围。

而对应到存储的具体实现方面，MergeTree也并不是一股脑地将数据直接写入.bin文件，而是经过了一番精心设计：首先，数据是经过压缩的，目前支持LZ4、ZSTD、Multiple和Delta几种算法，默认使用LZ4算法；其次，数据会事先依照ORDER BY的声明排序；最后，数据是以压缩数据块的形式被组织并写入.bin文件中的。

压缩数据块就好比一本书的文字段落，是组织文字的基本单元。这个概念十分重要，值得多花些篇幅进一步展开说明。

### 压缩数据块

一个压缩数据块由头信息和压缩数据两部分组成。头信息固定使用9位字节表示，具体由1个UInt8（1字节）整型和2个UInt32（4字节）整型组成，分别代表使用的压缩算法类型、压缩后的数据大小和压缩前的数据大小，具体如图6-14所示。

![压缩数据块示意图](images/%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E5%9D%97%E7%A4%BA%E6%84%8F%E5%9B%BE.png)

从图6-14所示中能够看到，.bin压缩文件是由多个压缩数据块组成的，而每个压缩数据块的头信息则是基于CompressionMethod_CompressedSize_UncompressedSize公式生成的。

通过ClickHouse提供的clickhouse-compressor工具，能够查询某个.bin文件中压缩数据的统计信息。以测试数据集hits_v1为例，执行下面的命令：

```shell
clickhouse-compressor --stat < /chbase/ /data/default/hits_v1/201403_1_34_3/JavaEnable.bin
```

执行后，会看到如下信息：

```
65536 12000
65536 14661
65536 4936
65536 7506
省略…
```

其中每一行数据代表着一个压缩数据块的头信息，其分别表示该压缩块中未压缩数据大小和压缩后数据大小（打印信息与物理存储的顺序刚好相反）。

每个压缩数据块的体积，按照其压缩前的数据字节大小，都被严格控制在64KB～1MB，其上下限分别由min_compress_block_size（默认65536）与max_compress_block_size（默认1048576）参数指定。而一个压缩数据块最终的大小，则和一个间隔（index_granularity）内数据的实际大小相关（是的，没错，又见到索引粒度这个老朋友了）。

MergeTree在数据具体的写入过程中，会依照索引粒度（默认情况下，每次取8192行），按批次获取数据并进行处理。如果把一批数据的未压缩大小设为size，则整个写入过程遵循以下规则：

1. **单个批次数据size<64KB** ：如果单个批次数据小于64KB，则继续获取下一批数据，直至累积到size>=64KB时，生成下一个压缩数据块。
2. **单个批次数据64KB<=size<=1MB** ：如果单个批次数据大小恰好在64KB与1MB之间，则直接生成下一个压缩数据块。
3. **单个批次数据size>1MB** ：如果单个批次数据直接超过1MB，则首先按照1MB大小截断并生成下一个压缩数据块。剩余数据继续依照上述规则执行。此时，会出现一个批次数据生成多个压缩数据块的情况。

整个过程逻辑如图6-15所示。

![切割压缩数据块的逻辑示意图](images/%E5%88%87%E5%89%B2%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E5%9D%97%E7%9A%84%E9%80%BB%E8%BE%91%E7%A4%BA%E6%84%8F%E5%9B%BE.png)

经过上述的介绍后我们知道，一个.bin文件是由1至多个压缩数据块组成的，每个压缩块大小在64KB～1MB之间。多个压缩数据块之间，按照写入顺序首尾相接，紧密地排列在一起，如图6-16所示。

在.bin文件中引入压缩数据块的目的至少有以下两个：其一，虽然数据被压缩后能够有效减少数据大小，降低存储空间并加速数据传输效率，但数据的压缩和解压动作，其本身也会带来额外的性能损耗。所以需要控制被压缩数据的大小，以求在性能损耗和压缩率之间寻求一种平衡。其二，在具体读取某一列数据时（.bin文件），首先需要将压缩数据加载到内存并解压，这样才能进行后续的数据处理。通过压缩数据块，可以在不读取整个.bin文件的情况下将读取粒度降低到压缩数据块级别，从而进一步缩小数据读取的范围。

![读取粒度精确到压缩数据块](images/%E8%AF%BB%E5%8F%96%E7%B2%92%E5%BA%A6%E7%B2%BE%E7%A1%AE%E5%88%B0%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E5%9D%97.png)

## 数据标记

如果把MergeTree比作一本书，primary.idx一级索引好比这本书的一级章节目录，.bin文件中的数据好比这本书中的文字，那么数据标记(.mrk)会为一级章节目录和具体的文字之间建立关联。对于数据标记而言，它记录了两点重要信息：其一，是一级章节对应的页码信息；其二，是一段文字在某一页中的起始位置信息。这样一来，通过数据标记就能够很快地从一本书中立即翻到关注内容所在的那一页，并知道从第几行开始阅读。



### 数据标记的生成规则

数据标记作为衔接一级索引和数据的桥梁，其像极了做过标记小抄的书签，而且书本中每个一级章节都拥有各自的书签。它们之间的关系如图6-17所示。

![图6-17　通过索引下标编号找到对应的数据标记](images/%E5%9B%BE6-17%E3%80%80%E9%80%9A%E8%BF%87%E7%B4%A2%E5%BC%95%E4%B8%8B%E6%A0%87%E7%BC%96%E5%8F%B7%E6%89%BE%E5%88%B0%E5%AF%B9%E5%BA%94%E7%9A%84%E6%95%B0%E6%8D%AE%E6%A0%87%E8%AE%B0.png)



从图6-17中一眼就能发现数据标记的首个特征，即数据标记和索引区间是对齐的，均按照index_granularity的粒度间隔。如此一来，只需简单通过索引区间的下标编号就可以直接找到对应的数据标记。

为了能够与数据衔接，数据标记文件也与.bin文件一一对应。即每一个列字段[Column].bin文件都有一个与之对应的[Column].mrk数据标记文件，用于记录数据在.bin文件中的偏移量信息。

一行标记数据使用一个元组表示，元组内包含两个整型数值的偏移量信息。它们分别表示在此段数据区间内，在对应的.bin压缩文件中，压缩数据块的起始偏移量；以及将该数据压缩块解压后，其未压缩数据的起始偏移量。图6-18所示是.mrk文件内标记数据的示意。

![图6-18　标记数据示意图](images/%E5%9B%BE6-18%E3%80%80%E6%A0%87%E8%AE%B0%E6%95%B0%E6%8D%AE%E7%A4%BA%E6%84%8F%E5%9B%BE.png)

如图6-18所示，每一行标记数据都表示了一个片段的数据（默认8192行）在.bin压缩文件中的读取位置信息。标记数据与一级索引数据不同，它并不能常驻内存，而是使用LRU（最近最少使用）缓存策略加快其取用速度。

### 数据标记的工作方式

MergeTree在读取数据时，必须通过标记数据的位置信息才能够找到所需要的数据。整个查找过程大致可以分为读取压缩数据块和读取数据两个步骤。为了便于解释，这里继续使用测试表hits_v1中的真实数据进行说明。图6-19所示为hits_v1测试表的JavaEnable字段及其标记数据与压缩数据的对应关系。

![图6-19　JavaEnable字段的标记文件和压缩数据文件的对应关系](images/%E5%9B%BE6-19%E3%80%80JavaEnable%E5%AD%97%E6%AE%B5%E7%9A%84%E6%A0%87%E8%AE%B0%E6%96%87%E4%BB%B6%E5%92%8C%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E6%96%87%E4%BB%B6%E7%9A%84%E5%AF%B9%E5%BA%94%E5%85%B3%E7%B3%BB.png)

首先，对图6-19所示左侧的标记数据做一番解释说明。JavaEnable字段的数据类型为UInt8，所以每行数值占用1字节。而hits_v1数据表的index_granularity粒度为8192，所以一个索引片段的数据大小恰好是8192B。按照6.5.2节介绍的压缩数据块的生成规则，如果单个批次数据小于64KB，则继续获取下一批数据，直至累积到size>=64KB时，生成下一个压缩数据块。因此在JavaEnable的标记文件中，每8行标记数据对应1个压缩数据块（1B*8192=8192B,64KB=65536B,65536/8192=8）。所以，从图6-19所示中能够看到，其左侧的标记数据中，8行数据的压缩文件偏移量都是相同的，因为这8行标记都指向了同一个压缩数据块。而在这8行的标记数据中，它们的解压缩数据块中的偏移量，则依次按照8192B（每行数据1B，每一个批次8192行数据）累加，当累加达到65536(64KB)时则置0。因为根据规则，此时会生成下一个压缩数据块。

理解了上述标记数据之后，接下来就开始介绍MergeTree具体是如何定位压缩数据块并读取数据的。

1. 读取压缩数据块： 在查询某一列数据时，MergeTree无须一次性加载整个.bin文件，而是可以根据需要，只加载特定的压缩数据块。而这项特性需要借助标记文件中所保存的压缩文件中的偏移量。

   在图6-19所示的标记数据中，上下相邻的两个压缩文件中的起始偏移量，构成了与获取当前标记对应的压缩数据块的偏移量区间。由当前标记数据开始，向下寻找，直到找到不同的压缩文件偏移量为止。此时得到的一组偏移量区间即是压缩数据块在.bin文件中的偏移量。例如在图6-19所示中，读取右侧.bin文件中[0，12016]字节数据，就能获取第0个压缩数据块。

   细心的读者可能会发现，在.mrk文件中，第0个压缩数据块的截止偏移量是12016。而在.bin数据文件中，第0个压缩数据块的压缩大小是12000。为什么两个数值不同呢？其实原因很简单，12000只是数据压缩后的字节数，并没有包含头信息部分。而一个完整的压缩数据块是由头信息加上压缩数据组成的，它的头信息固定由9个字节组成，压缩后大小为8个字节。所以，12016=8+12000+8，其定位方法如图6-19右上角所示。压缩数据块被整个加载到内存之后，会进行解压，在这之后就进入具体数据的读取环节了。

   

2. 读取数据： 在读取解压后的数据时，MergeTree并不需要一次性扫描整段解压数据，它可以根据需要，以index_granularity的粒度加载特定的一小段。为了实现这项特性，需要借助标记文件中保存的解压数据块中的偏移量。

同样的，在图6-19所示的标记数据中，上下相邻两个解压缩数据块中的起始偏移量，构成了与获取当前标记对应的数据的偏移量区间。通过这个区间，能够在它的压缩块被解压之后，依照偏移量按需读取数据。例如在图6-19所示中，通过[0，8192]能够读取压缩数据块0中的第一个数据片段

## 对于分区、索引、标记和压缩数据的协同总结

分区、索引、标记和压缩数据，就好比是MergeTree给出的一套组合拳，使用恰当时威力无穷。那么，在依次介绍了各自的特点之后，现在将它们聚在一块进行一番总结。接下来，就分别从写入过程、查询过程，以及数据标记与压缩数据块的三种对应关系的角度展开介绍。

### 写入过程

数据写入的第一步是生成分区目录，伴随着每一批数据的写入，都会生成一个新的分区目录。在后续的某一时刻，属于相同分区的目录会依照规则合并到一起；接着，按照index_granularity索引粒度，会分别生成primary.idx一级索引（如果声明了二级索引，还会创建二级索引文件）、每一个列字段的.mrk数据标记和.bin压缩数据文件。图6-20所示是一张MergeTree表在写入数据时，它的分区目录、索引、标记和压缩数据的生成过程。

![图6-20　分区目录、索引、标记和压缩数据的生成过程示意](images/%E5%9B%BE6-20%E3%80%80%E5%88%86%E5%8C%BA%E7%9B%AE%E5%BD%95%E3%80%81%E7%B4%A2%E5%BC%95%E3%80%81%E6%A0%87%E8%AE%B0%E5%92%8C%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E7%9A%84%E7%94%9F%E6%88%90%E8%BF%87%E7%A8%8B%E7%A4%BA%E6%84%8F.png)

从分区目录201403_1_34_3能够得知，该分区数据共分34批写入，期间发生过3次合并。在数据写入的过程中，依据index_granularity的粒度，依次为每个区间的数据生成索引、标记和压缩数据块。其中，索引和标记区间是对齐的，而标记与压缩块则根据区间数据大小的不同，会生成多对一、一对一和一对多三种关系。

### 查询过程

数据查询的本质，可以看作一个不断减小数据范围的过程。在最理想的情况下，MergeTree首先可以依次借助分区索引、一级索引和二级索引，将数据扫描范围缩至最小。然后再借助数据标记，将需要解压与计算的数据范围缩至最小。以图6-21所示为例，它示意了在最优的情况下，经过层层过滤，最终获取最小范围数据的过程。

![图6-21　将扫描数据范围最小化的过程](images/%E5%9B%BE6-21%E3%80%80%E5%B0%86%E6%89%AB%E6%8F%8F%E6%95%B0%E6%8D%AE%E8%8C%83%E5%9B%B4%E6%9C%80%E5%B0%8F%E5%8C%96%E7%9A%84%E8%BF%87%E7%A8%8B.png)

如果一条查询语句没有指定任何WHERE条件，或是指定了WHERE条件，但条件没有匹配到任何索引（分区索引、一级索引和二级索引），那么MergeTree就不能预先减小数据范围。在后续进行数据查询时，它会扫描所有分区目录，以及目录内索引段的最大区间。虽然不能减少数据范围，但是MergeTree仍然能够借助数据标记，以多线程的形式同时读取多个压缩数据块，以提升性能。

### 数据标记与压缩数据块的对应关系

由于压缩数据块的划分，与一个间隔（index_granularity）内的数据大小相关，每个压缩数据块的体积都被严格控制在64KB～1MB。而一个间隔（index_granularity）的数据，又只会产生一行数据标记。那么根据一个间隔内数据的实际字节大小，数据标记和压缩数据块之间会产生三种不同的对应关系。接下来使用具体示例做进一步说明，对于示例数据，仍然是测试表hits_v1，其中index_granularity粒度为8192，数据总量为8873898行。

#### 1.多对一

多个数据标记对应一个压缩数据块，当一个间隔（index_granularity）内的数据未压缩大小size小于64KB时，会出现这种对应关系。

以hits_v1测试表的JavaEnable字段为例。JavaEnable数据类型为UInt8，大小为1B，则一个间隔内数据大小为8192B。所以在此种情形下，每8个数据标记会对应同一个压缩数据块，如图6-22所示。

![图6-22　多个数据标记对应同一个压缩数据块的示意](images/%E5%9B%BE6-22%E3%80%80%E5%A4%9A%E4%B8%AA%E6%95%B0%E6%8D%AE%E6%A0%87%E8%AE%B0%E5%AF%B9%E5%BA%94%E5%90%8C%E4%B8%80%E4%B8%AA%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E5%9D%97%E7%9A%84%E7%A4%BA%E6%84%8F.png)

#### 2.一对一

一个数据标记对应一个压缩数据块，当一个间隔（index_granularity）内的数据未压缩大小size大于等于64KB且小于等于1MB时，会出现这种对应关系。

以hits_v1测试表的URLHash字段为例。URLHash数据类型为UInt64，大小为8B，则一个间隔内数据大小为65536B，恰好等于64KB。所以在此种情形下，数据标记与压缩数据块是一对一的关系，如图6-23所示。

![图6-23　一个数据标记对应一个压缩数据块的示意](images/%E5%9B%BE6-23%E3%80%80%E4%B8%80%E4%B8%AA%E6%95%B0%E6%8D%AE%E6%A0%87%E8%AE%B0%E5%AF%B9%E5%BA%94%E4%B8%80%E4%B8%AA%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E5%9D%97%E7%9A%84%E7%A4%BA%E6%84%8F.png)

#### 3.一对多

一个数据标记对应多个压缩数据块，当一个间隔（index_granularity）内的数据未压缩大小size直接大于1MB时，会出现这种对应关系。

以hits_v1测试表的URL字段为例。URL数据类型为String，大小根据实际内容而定。如图6-24所示，编号45的标记对应了2个压缩数据块。

![图6-24　一个数据标记对应多个压缩数据块的示意](images/%E5%9B%BE6-24%E3%80%80%E4%B8%80%E4%B8%AA%E6%95%B0%E6%8D%AE%E6%A0%87%E8%AE%B0%E5%AF%B9%E5%BA%94%E5%A4%9A%E4%B8%AA%E5%8E%8B%E7%BC%A9%E6%95%B0%E6%8D%AE%E5%9D%97%E7%9A%84%E7%A4%BA%E6%84%8F.png)

## 本章小结

本章全方面、立体地解读了MergeTree表引擎的工作原理：首先，解释了MergeTree的基础属性和物理存储结构；接着，依次介绍了数据分区、一级索引、二级索引、数据存储和数据标记的重要特性；最后，结合实际样例数据，进一步总结了MergeTree上述特性在一起协同时的工作过程。掌握本章的内容，即掌握了合并树系列表引擎的精髓。下一章将进一步介绍MergeTree家族中其他常见表引擎的具体使用方法。
