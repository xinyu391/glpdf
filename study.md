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
字体编码与CID
CID和Unicode 
Adobe类有Japan1、Korea1、GB1、CNS1  4种字集
CID转Unicode字码表 名称后面加-UCS2 例如Adobe-CNS1-UCS2，
每个CMap输入字符映射范围有begincodespacerange指定， 原字符可以是1个字节，也可以式2个字节，如下：
2 begincodespacerange
	<00>   <80>
	<8140> <FEFE>
endcodespacerange


- IdentityCMap: A special case of CMap, where the _map array implicitly has a length of 65536 and each element is equal to its index.

http://www.hunterpro.net/?p=293

对于一个非Type0的字型，输入字码一律是一个byte，取得Unicode字码的方式如下：
(1) 如果有ToUnicode CMap的话，便直接到该CMap取得对应的Unicode字码
(2) 如果找不到对应的Unicode字码，便依输入字码到Encoding信息里取得该字码对应的字符名称
(3) 依照取得的字符名称转成对应的Unicode字码（不一定能转换，此时表示显示出来的字没有对应的Unicode字码，或是writer并未附相关的Unicode字码转换信息）

CIDSystemInfo里记录了该字型使用的字集，而每种字集的CID都是固定的，因此只要知道这个字集，即可将Encoding CMap转出来的CID再转成Unicode字码， 称后面加-UCS2的就是了，例如Adobe-CNS1-UCS2	


没有找到ToUnicode的话，就根据CIDSystemInfo中的 /Registry-/Ordering-/Supplement找到 charCode 转CID的cmap文件。
然后再找到Adobe-xxx-UCS2文件， 查找CID对应的unicode 码。


#### CMapy有两种类型
+ ToUnicode Map
用于char code查找对应unicode，文件中包含 beginbfrange,beginbfchar
+ encoding Map
用于char code查找cid，文件中包含 begincidrange，begincidchar

++http://bbs.csdn.net/topics/340109816	
Identity-H和Identity-V，这两个都是直接将2 byte输入字码视为CID处理
