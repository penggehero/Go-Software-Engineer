# Grcp 序列化入门-Varints 编码

Grpc使用Varints 作为基础的编码方式，先简单入门一下。

Varints are a method of serializing integers using one or more bytes. Smaller numbers take a smaller number of bytes.

Each byte in a varint, except the last byte, has the *most significant bit* (MSB) set – this indicates that there are further bytes to come. The lower 7 bits of each byte are used to store the two's complement representation of the number in groups of 7 bits, **least significant group first**.

So, for example, here is the number 1 – it's a single byte, so the MSB is not set:

Varints 是一种使用一个或多个字节序列化整数的方法。较小的数字占用较少的字节数。

varint 中的每个字节，除了最后一个字节，都设置了*最高有效位*(MSB)——这表明还有更多字节要到来。每个字节的低 7 位用于存储以 7 位为一组的数字的二进制补码表示，**最低有效组在前**。

例如，这里是数字 1——它是一个单字节，所以 MSB 没有设置：

```proto
0000 0001
```

这里是 300——这有点复杂：

```proto
1010 1100 0000 0010
```

你怎么知道这是300？首先，您从每个字节中删除 MSB，因为这只是为了告诉我们是否已经到达数字的末尾（如您所见，它设置在第一个字节中，因为 varint 中有多个字节） ：

```proto
 1010 1100 0000 0010
→ 010 1100  000 0010
```

您颠倒了两组 7 位，因为 varint 首先存储具有最低有效组的数字。然后将它们连接起来以获得最终值：

```proto
000 0010  010 1100
→  000 0010 ++ 010 1100
→  100101100
→  256 + 32 + 8 + 4 = 300
```

Varints 是一种将整数压缩到比通常需要的更小的空间的方法。默认情况下，出于硬件效率的原因，计算机使用固定长度的整数。但是，在传输或存储整数时，将它们压缩以节省带宽很重要。Google 的 Protobuf使用后一种技术，使用每个字节的最高位来指示是否有更多字节到来。

Varints 基于大多数数字不是均匀分布的想法。几乎总是，较小的数字在计算中比较大的数字更常见。varints 的权衡是在较大的数字上花费更多的位，在较小的数字上花费更少的位。例如，一个几乎总是小于 256 的 64 位整数将浪费固定宽度表示的前 56 位。编码 varint 的两种常用方法是长度前缀和连续位。



只需 2 个字节，我们就可以对数字 300 进行编码。请注意每个字节的最高位如何判断是否还有更多字节。如果最高位是 1，我们知道继续查找。如果它是 0，我们知道它是最后一个字节。解码以相反的顺序进行：删除最高位，反转组的顺序，然后连接位以获取原始数字。

这种技术真的很强大，因为它可以编码任意大小的数字！使用 32 或 64 位数字，您将自己限制在最大值。使用 varint，您可以将大数和小数一起编码，即使它们最初都是大数。更酷的是，varints 可以连接起来！因为总是很清楚一个数字在哪里结束，另一个数字在哪里开始，所以可以按顺序写出来。Varints 是自定界的。

#### 优点：

- 字节对齐。无需将编码数字填充到字节边界。
- 高效率。一个 64 位数字最多可以编码为 10 个字节。

伪代码实现：

```go
func VarintsEncoding(b int) (res []byte) {
   sprintf := fmt.Sprintf("%b", b)
   sprintf = FillZero(sprintf)
   chunks := Reverse(Chunks(sprintf, 7))
   for i := range chunks {
      if i == len(chunks)-1 {
         chunks[i] = "0" + chunks[i]
      } else {
         chunks[i] = "1" + chunks[i]
      }
      parseInt, err := strconv.ParseInt(chunks[i], 2, 64)
      if err != nil {
         panic(err)
      }
      res = append(res, byte(parseInt))
   }
   fmt.Println(strings.Join(chunks," "))
   return
}

func FillZero(s string) string {
   n := len(s) % 7
   if n == 0 {
      return s
   }
   var rns []rune
   for i := 0; i < 7-n; i++ {
      rns = append(rns, '0')
   }
   rns = append(rns, []rune(s)...)
   return string(rns)
}

func Chunks(s string, chunkSize int) []string {
   if len(s) == 0 {
      return nil
   }
   if chunkSize >= len(s) {
      return []string{s}
   }
   var chunks = make([]string, 0, (len(s)-1)/chunkSize+1)
   currentLen := 0
   currentStart := 0
   for i := range s {
      if currentLen == chunkSize {
         chunks = append(chunks, s[currentStart:i])
         currentLen = 0
         currentStart = i
      }
      currentLen++
   }
   chunks = append(chunks, s[currentStart:])
   return chunks
}

func Reverse(s []string) []string {
   for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
      s[i], s[j] = s[j], s[i]
   }
   return s
}
```



参考：

https://carlmastrangelo.com/blog/lets-make-a-varint

https://developers.google.com/protocol-buffers/docs/encoding

https://en.wikipedia.org/wiki/Variable-length_quantity
