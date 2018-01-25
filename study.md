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
Adobe类有Japan1、Korea1、GB1、CNS1  4种字集
CID转Unicode字码表 名称后面加-UCS2 例如Adobe-CNS1-UCS2，

http://www.hunterpro.net/?p=293

对于一个非Type0的字型，输入字码一律是一个byte，取得Unicode字码的方式如下：
(1) 如果有ToUnicode CMap的话，便直接到该CMap取得对应的Unicode字码
(2) 如果找不到对应的Unicode字码，便依输入字码到Encoding信息里取得该字码对应的字符名称
(3) 依照取得的字符名称转成对应的Unicode字码（不一定能转换，此时表示显示出来的字没有对应的Unicode字码，或是writer并未附相关的Unicode字码转换信息）

#### CMapy有两种类型
+ ToUnicode Map
用于char code查找对应unicode，文件中包含 beginbfrange,beginbfchar
+ encoding Map
用于char code查找cid，文件中包含 begincidrange，begincidchar
