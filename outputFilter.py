
class Analyzer:

    def __init__(self,filename = "out.txt"):

        self.lines = []
        with open(filename,"r") as f:
            self.lines = f.readlines()
        
    
    def show(self):
        for line in self.lines:
            print(line)
    
    def len(self):
        return len(self.lines)
    
    def filter(self,value):
        i = 0
        ans = []
        while i < len(self.lines):
            if value in self.lines[i]:
                ans += [self.lines[i]]
            i = i + 1
        self.lines = ans
    
    def store(self,fname):
        with open(fname,'w') as f:
            for line in self.lines:
                f.write(line)
    


import sys

if __name__ == "__main__":
    a = Analyzer(filename = sys.argv[1])
    print(a.len())
    # a.filter('ssv-node-1')
    print(a.len())
    a.filter('$$$$$')
    # a.filter('starting')
    # a.filter('QBFT')
    # a.filter('instance')
    # a.filter(""""publicKey": "b3abb58d18a587ff17cdbd85be687622eb2add07e80ca5190745badb03a216f8969b59405c85bd5a188008bcdf5a58c5", "role": "SYNC_COMMITTEE", "height": 1050""")
    print(a.len())
    a.store(sys.argv[2])
