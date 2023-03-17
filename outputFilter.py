
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
    



if __name__ == "__main__":
    a = Analyzer()
    print(a.len())
    # a.filter('ssv-node-1')
    print(a.len())
    a.filter('$$$$$$')
    print(a.len())
    a.store('out_filtered.txt')

