## PDF Font and encoding
pdf文件中的文本字符串，使用CID进行编码的。
使用cid编码，可以快速的从内嵌字体中查找自行。
解析pdf，需要一个cmap文件，来进行cid到unicode的映射。

cid 在不同语言集上对应的unicode字符是不一样的
不同语言有不同的cmap文件
在解析过程中，需要解析到cmap的名称，然后在这个cmap文件中查找cid对应的unicode字符

### 
+ 对于<>字符串，解析时，不进行转码（hex字符到byte），具体用到时在进行转换
cmap:maps character codes to glyph selector(cid)