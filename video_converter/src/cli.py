from typing import Callable, TypedDict
from os import getcwd

class Params(TypedDict):
    out: str
    prefix: str
    index: int
    height: int
    kbps: int
    fps: int

class Cli:
    out: Union[str, None]
    prefix: Union[str, None]
    index: Union[int, None]
    height: Union[int, None]
    kbps: Union[int, None]
    fps: Union[int, None]

    def __init__(self):
        self.out: str = f"{getcwd()}/dest"
        self.prefix: str = "video"
        self.index: int = 0  
        self.height: int = 180
        self.kbps: int = 150
        self.fps: int = 15

    def parse(self, key: str, value: str) -> None:
        if key == '-p':
            self.prefix = value
            return

        if key == '-i':
            self.index = value
            return

        if key == '-h':
            self.height = value
            return

        if key == '-k':
            self.kbps = value
            return

        if key == '-f':
            self.fps = value
            return

        if key == '-d':
            if value.endswith('/'):
                raise Exception("Dest folder should not end with /")
            
            self.out = value
            return

        if key == '--help':
            self.help()
            return


    def help(self) -> None:
        print("""

        option usage: [-flag=value]        

        Options:
        -p prefix for name (default 'video')
        -i starting index for prefix (default '0')
        -h video height (default 180)
        -k kbps (default 150)
        -f fps (default 15)
        -d output destination (default dest/ folder)
        """)
    
    def get_params(self) -> Params:
        return Params(
            prefix=self.prefix,
            index=self.index,
            height=self.height,
            kbps=self.kbps,
            fps=self.fps,
            out=self.out
        )
