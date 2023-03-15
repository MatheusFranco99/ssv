
def extract(line,field):
    return line.split(field)[1].split()[0][:-1]


def extractTime(line):
    return int(line.split("\"time(micro)\":")[1].split()[0][:-1])


class UponChangeRound:


    def __init__(self,sender,round):
        self.sender = sender
        self.round = round
        self.start = None
        self.finish = None
        self.hasBroadcast = False
        self.broadcastStart = None
        self.broadcastFinish = None
        self.showedSingleton = False
    
    def process(self,line):
        if "broadcast start" in line:
            self.hasBroadcast = True
            self.broadcastStart = extractTime(line)
        elif "broadcast finish" in line:
            self.broadcastFinish = extractTime(line)
        elif "UponRoundChange start" in line:
            self.start = extractTime(line)
        elif "UponRoundChange return" in line:
            self.finish = extractTime(line)
    
    def len(self):
        return self.finish - self.start
    
    def show(self):
        print(f"UponRoundChange {self.sender=} {self.round=}")
        print(f"\t{self.start=}")
        t = self.start
        if self.hasBroadcast:
            print(f"\t{self.broadcastStart=}, delta:{self.broadcastStart-t}")
        if self.broadcastFinish != None:    
            print(f"\t{self.broadcastFinish=}, delta:{self.broadcastFinish-t}")
        if self.finish != None:
            print(f"\t{self.finish=} delta:{self.finish-t}")
            print(f"total: {self.len()}")
    
    def showSingleton(self,t0):
        if not self.showedSingleton:
            self.showedSingleton = True
            print(f"UponRoundChange {self.sender=} {self.round=}")
            print(f"\t{self.start=}, time: {self.start - t0}")
            t = self.start
            if self.hasBroadcast:
                print(f"\t{self.broadcastStart=}, time: {self.broadcastStart - t0}, delta:{self.broadcastStart-t}")
            if self.broadcastFinish != None:    
                print(f"\t{self.broadcastFinish=}, time: {self.broadcastFinish - t0}, delta:{self.broadcastFinish-t}")
            if self.finish != None:
                print(f"\t{self.finish=}, time: {self.finish - t0}, delta:{self.finish-t}")
                print(f"total: {self.len()}")



class UponChangeRoundController:

    def __init__(self):
        self.entries = {}
    
    def getAll(self):
        ans = []
        for sender in self.entries:
            for round in self.entries[sender]:
                ans += [self.entries[sender][round]]
        return ans

    def get(self,sender,round):
        try:
            return self.entries[sender][round]
        except:
            return None
    
    def add(self,sender,round):
        if sender not in self.entries:
            self.entries[sender] = {}
        
        if round in self.entries[sender]:
            return self.entries[sender][round]
        else:
            up = UponChangeRound(sender,round)
            self.entries[sender][round] = up
            return up





class UponRoundTimeout:


    def __init__(self,round):
        self.round = round
        self.start = None
        self.finish = None
        self.hasBroadcast = False
        self.broadcastStart = None
        self.broadcastFinish = None
        self.showedSingleton = False
    
    def process(self,line):
        if "broadcast start" in line:
            self.hasBroadcast = True
            self.broadcastStart = extractTime(line)
        elif "broadcast finish" in line:
            self.broadcastFinish = extractTime(line)
        elif "UponRoundTimeout start" in line:
            self.start = extractTime(line)
        elif "UponRoundTimeout return" in line:
            self.finish = extractTime(line)
    
    def len(self):
        return self.finish - self.start
    
    def show(self):
        print(f"UponRoundTimeout {self.round=}")
        print(f"\t{self.start=}")
        t = self.start
        if self.hasBroadcast:
            print(f"\t{self.broadcastStart=}, delta:{self.broadcastStart-t}")
        if self.broadcastFinish != None:    
            print(f"\t{self.broadcastFinish=}, delta:{self.broadcastFinish-t}")
        if self.finish != None:
            print(f"\t{self.finish=} delta:{self.finish-t}")
            print(f"total: {self.len()}")
    
    def showSingleton(self,t0):
        if not self.showedSingleton:
            self.showedSingleton = True
            print(f"UponRoundTimeout {self.round=}")
            print(f"\t{self.start=}, time: {self.start - t0}")
            t = self.start
            if self.hasBroadcast:
                print(f"\t{self.broadcastStart=}, time: {self.broadcastStart - t0}, delta:{self.broadcastStart-t}")
            if self.broadcastFinish != None:    
                print(f"\t{self.broadcastFinish=}, time: {self.broadcastFinish - t0}, delta:{self.broadcastFinish-t}")
            if self.finish != None:
                print(f"\t{self.finish=}, time: {self.finish - t0}, delta:{self.finish-t}")
                print(f"total: {self.len()}")


class UponRoundTimeoutController:

    def __init__(self):
        self.entries = {}
    
    def getAll(self):
        ans = []
        for round in self.entries:
            ans += [self.entries[round]]
        return ans

    def get(self,round):
        try:
            return self.entries[round]
        except:
            return None
    
    def add(self,round):
        if round in self.entries:
            return self.entries[round]
        else:
            up = UponRoundTimeout(round)
            self.entries[round] = up
            return up




class UponCommit:

    def __init__(self,sender,round):
        self.sender = sender
        self.round = round
        self.start = None
        self.finish = None
        self.hasAggregate = False
        self.aggregateStart = None
        self.aggregateFinish = None
        self.showedSingleton = False
    
    def process(self,line):
        if "start aggregate" in line:
            self.hasAggregate = True
            self.aggregateStart = extractTime(line)
        elif "finish aggregate" in line:
            self.aggregateFinish = extractTime(line)
        elif "UponCommit start" in line:
            self.start = extractTime(line)
        elif "UponCommit return" in line:
            self.finish = extractTime(line)
    
    def len(self):
        return self.finish - self.start
    
    def show(self):
        print(f"UponCommit {self.sender=} {self.round=}")
        print(f"\t{self.start=}")
        t = self.start
        if self.hasAggregate:
            print(f"\t{self.aggregateStart=}, delta:{self.aggregateStart-t}")
        if self.aggregateFinish != None:    
            print(f"\t{self.aggregateFinish=}, delta:{self.aggregateFinish-t}")
        if self.finish != None:
            print(f"\t{self.finish=} delta:{self.finish-t}")
            print(f"total: {self.len()}")

    def showSingleton(self,t0):
        if not self.showedSingleton:
            self.showedSingleton = True
            print(f"UponCommit {self.sender=} {self.round=}")
            print(f"\t{self.start=}, time: {self.start - t0}")
            t = self.start
            if self.hasAggregate:
                print(f"\t{self.aggregateStart=}, time: {self.aggregateStart - t0}, delta:{self.aggregateStart-t}")
            if self.aggregateFinish != None:    
                print(f"\t{self.aggregateFinish=}, time: {self.aggregateFinish - t0}, delta:{self.aggregateFinish-t}")
            if self.finish != None:
                print(f"\t{self.finish=}, time: {self.finish - t0}, delta:{self.finish-t}")
                print(f"total: {self.len()}")


class UponCommitController:

    def __init__(self):
        self.entries = {}
    
    def getAll(self):
        ans = []
        for sender in self.entries:
            for round in self.entries[sender]:
                ans += [self.entries[sender][round]]
        return ans

    def get(self,sender,round):
        try:
            return self.entries[sender][round]
        except:
            return None
    
    def add(self,sender,round):
        if sender not in self.entries:
            self.entries[sender] = {}
        
        if round in self.entries[sender]:
            return self.entries[sender][round]
        else:
            up = UponCommit(sender,round)
            self.entries[sender][round] = up
            return up




class UponPrepare:

    def __init__(self,sender,round):
        self.sender = sender
        self.round = round
        self.start = None
        self.finish = None
        self.hasBroadcast = False
        self.broadcastStart = None
        self.broadcastFinish = None
        self.showedSingleton = False
    
    def process(self,line):
        if "broadcast start" in line:
            self.hasBroadcast = True
            self.broadcastStart = extractTime(line)
        elif "broadcast finish" in line:
            self.broadcastFinish = extractTime(line)
        elif "UponPrepare start" in line:
            self.start = extractTime(line)
        elif "UponPrepare return" in line:
            self.finish = extractTime(line)
    
    def len(self):
        return self.finish - self.start
    
    def show(self):
        print(f"UponPrepare {self.sender=} {self.round=}")
        print(f"\t{self.start=}")
        t = self.start
        if self.hasBroadcast:
            print(f"\t{self.broadcastStart=}, delta:{self.broadcastStart-t}")
        if self.broadcastFinish != None:    
            print(f"\t{self.broadcastFinish=}, delta:{self.broadcastFinish-t}")
        if self.finish != None:
            print(f"\t{self.finish=} delta:{self.finish-t}")
            print(f"total: {self.len()}")
    
    def showSingleton(self,t0):
        if not self.showedSingleton:
            self.showedSingleton = True
            print(f"UponPrepare {self.sender=} {self.round=}")
            print(f"\t{self.start=}, time: {self.start - t0}")
            t = self.start
            if self.hasBroadcast:
                print(f"\t{self.broadcastStart=}, time: {self.broadcastStart - t0}, delta:{self.broadcastStart-t}")
            if self.broadcastFinish != None:    
                print(f"\t{self.broadcastFinish=}, time: {self.broadcastFinish - t0}, delta:{self.broadcastFinish-t}")
            if self.finish != None:
                print(f"\t{self.finish=}, time: {self.finish - t0}, delta:{self.finish-t}")
                print(f"total: {self.len()}")


class UponPrepareController:

    def __init__(self):
        self.entries = {}
    
    def getAll(self):
        ans = []
        for sender in self.entries:
            for round in self.entries[sender]:
                ans += [self.entries[sender][round]]
        return ans


    def get(self,sender,round):
        try:
            return self.entries[sender][round]
        except:
            return None
    
    def add(self,sender,round):
        if sender not in self.entries:
            self.entries[sender] = {}
        
        if round in self.entries[sender]:
            return self.entries[sender][round]
        else:
            up = UponPrepare(sender,round)
            self.entries[sender][round] = up
            return up



class UponProposal:

    def __init__(self,sender,round):
        self.sender = sender
        self.round = round
        self.start = None
        self.finish = None
        self.hasBroadcast = False
        self.broadcastStart = None
        self.broadcastFinish = None
        self.showedSingleton = False
    
    def process(self,line):
        if "broadcast start" in line:
            self.hasBroadcast = True
            self.broadcastStart = extractTime(line)
        elif "broadcast finish" in line:
            self.broadcastFinish = extractTime(line)
        elif "UponProposal start." in line:
            self.start = extractTime(line)
        elif "UponProposal return." in line:
            self.finish = extractTime(line)
    
    def len(self):
        return self.finish - self.start
    
    def show(self):
        print(f"UponProposal {self.sender=} {self.round=}")
        print(f"\t{self.start=}")
        t = self.start
        if self.hasBroadcast:
            print(f"\t{self.broadcastStart=}, delta:{self.broadcastStart-t}")
        if self.broadcastFinish != None:    
            print(f"\t{self.broadcastFinish=}, delta:{self.broadcastFinish-t}")
        if self.finish != None:
            print(f"\t{self.finish=} delta:{self.finish-t}")
            print(f"total: {self.len()}")
    
    def showSingleton(self,t0):
        if not self.showedSingleton:
            self.showedSingleton = True
            print(f"UponProposal {self.sender=} {self.round=}")
            print(f"\t{self.start=}, time: {self.start - t0}")
            t = self.start
            if self.hasBroadcast:
                print(f"\t{self.broadcastStart=}, time: {self.broadcastStart - t0}, delta:{self.broadcastStart-t}")
            if self.broadcastFinish != None:    
                print(f"\t{self.broadcastFinish=}, time: {self.broadcastFinish - t0}, delta:{self.broadcastFinish-t}")
            if self.finish != None:
                print(f"\t{self.finish=}, time: {self.finish - t0}, delta:{self.finish-t}")
                print(f"total: {self.len()}")


class UponProposalController:

    def __init__(self):
        self.uponProposals = {}
    
    def getAll(self):
        ans = []
        for sender in self.uponProposals:
            for round in self.uponProposals[sender]:
                ans += [self.uponProposals[sender][round]]
        return ans

    def get(self,sender,round):
        try:
            return self.uponProposals[sender][round]
        except:
            return None
    
    def add(self,sender,round):
        if sender not in self.uponProposals:
            self.uponProposals[sender] = {}
        
        if round in self.uponProposals[sender]:
            return self.uponProposals[sender][round]
        else:
            up = UponProposal(sender,round)
            self.uponProposals[sender][round] = up
            return up


class Execution:

    def __init__(self,publicKey = None, role = None, height = None):
        self.publicKey = publicKey
        self.role = role
        self.height = height
        self.lines = []
    
    def add(self,line):
        self.lines.append(line)
    
    def show(self):
        print(f"{self.publicKey=},{self.role=},{self.height=}")
        for line in self.lines:
            print(line)

    def analyse(self):
        self.show()

        uponProposalController = UponProposalController()
        uponPrepareController = UponPrepareController()
        uponCommitController = UponCommitController()
        uponChangeRoundController = UponChangeRoundController()
        uponRoundTimeoutController = UponRoundTimeoutController()

        for line in self.lines:
            if "UponProposal" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponProposalController.add(sender,round)
                up.process(line)
            if "UponPrepare" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponPrepareController.add(sender,round)
                up.process(line)
            if "UponCommit" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponCommitController.add(sender,round)
                up.process(line)
            if "UponRoundChange" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponChangeRoundController.add(sender,round)
                up.process(line)
            if "UponRoundTimeout" in line:
                round = extract(line,"\"round\":")
                up = uponRoundTimeoutController.add(round)
                up.process(line)
            
        # for up in uponProposalController.getAll():
        #     up.show()
        # for up in uponPrepareController.getAll():
        #     up.show()
        # for up in uponCommitController.getAll():
        #     up.show()
        # for up in uponChangeRoundController.getAll():
        #     up.show()
        # for up in uponRoundTimeoutController.getAll():
        #     up.show()

        self.start = None
        self.finish = None
        for line in self.lines:
            if "starting QBFT instance" in line:
                self.start = extractTime(line)
            elif "Decided on value with commit" in line:
                self.finish = extractTime(line)

        if self.start == None:
            return        

        for line in self.lines:
            if "UponProposal" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponProposalController.get(sender,round)
                if up != None:
                    up.showSingleton(self.start)
            if "UponPrepare" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponPrepareController.get(sender,round)
                if up != None:
                    up.showSingleton(self.start)
            if "UponCommit" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponCommitController.get(sender,round)
                if up != None:
                    up.showSingleton(self.start)
            if "UponRoundChange" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = uponChangeRoundController.get(sender,round)
                if up != None:
                    up.showSingleton(self.start)
            if "UponRoundTimeout" in line:
                round = extract(line,"\"round\":")
                up = uponRoundTimeoutController.get(round)
                if up != None:
                    up.showSingleton(self.start)
        
        if self.finish != None:
            print(f"Total duration: {self.finish - self.start}")
        
        


class ExecutionController:

    def __init__(self):
        self.executions = []
        self.publicKeyMap = {}

    
    def addExecution(self,publicKey = None, role = None, height = None):
        if publicKey not in self.publicKeyMap:
            self.publicKeyMap[publicKey] = {}
        if role not in self.publicKeyMap[publicKey]:
            self.publicKeyMap[publicKey][role] = {}
        
        if height not in self.publicKeyMap[publicKey][role]:
            execution = Execution(publicKey = publicKey, role = role, height = height)
            self.publicKeyMap[publicKey][role][height] = execution
            self.executions.append(execution)
            return execution
        else:
            return self.publicKeyMap[publicKey][role][height]


    def getExecution(self,publicKey,role,height):
        try:
            return self.publicKeyMap[publicKey][role][height]
        except:
            return None
        
    def hasExecution(self,publicKey,role,height):
        try:
            e = self.publicKeyMap[publicKey][role][height]
            return True
        except:
            return False

    def addToExecution(self,publicKey,role,height,line):
        if not self.hasExecution(publicKey,role,height):
            self.addExecution(publicKey,role,height)
        try:
            self.publicKeyMap[publicKey][role][height].add(line)
        except:
            return None
    
    def show(self):
        for pk in self.publicKeyMap:
            for role in self.publicKeyMap[pk]:
                for h in self.publicKeyMap[pk][role]:
                    print(f"{pk=},{role=},{h=}")
    
    def getAt(self,index):
        return self.executions[index]

    def getAll(self):
        return self.executions



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
    
    def test(self):
        # for i in range(10):
        #     print(self.lines[i])

        #     terms = ["\"publicKey\":","\"role\":","\"height\":"]
        #     for term in terms:
        #         if term in self.lines[i]:
        #             v = self.lines[i].split(term)[1].split()[0].split(",")[0]
        #             print(term,v)
        

        exec_ctrl = ExecutionController()

        def getExecutionParams(line):
            ans = {}
            terms = ["\"publicKey\":","\"role\":","\"height\":"]
            for term in terms:
                if term in line:
                    v = line.split(term)[1].split()[0].split(",")[0]
                    ans[term] = v
            return ans

        for line in self.lines:
            execParams = getExecutionParams(line)
            execution = exec_ctrl.addExecution(execParams["\"publicKey\":"],execParams["\"role\":"],execParams["\"height\":"])
            execution.add(line)

        return exec_ctrl

    def filter(self,value):
        i = 0
        ans = []
        while i < len(self.lines):
            if value in self.lines[i]:
                ans += [self.lines[i]]
            i = i + 1
        self.lines = ans

if __name__ == "__main__":
    a = Analyzer()
    a.filter('ssv-node-1')
    a.filter('$$$$$$')
    # a.filter("a7372f1be3e7bf1aa20000103a787929d2b884c2081392571cfc7f5034a91b79127706c6a2acde2c0277009fbd9943d1")
    # a.filter("\"height\": 0")
    exec_ctrl = a.test()
    # exec_ctrl.show()
    # execution = exec_ctrl.getAt(1)
    # execution.analyse()
    for e in exec_ctrl.getAll():
        e.analyse()
    # a.show()
    print(a.len())