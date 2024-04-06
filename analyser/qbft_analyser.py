
import plotly.express as px
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from datetime import datetime, timedelta
from matplotlib.patches import Patch
import statistics

import argparse
from logger import Logger
from colors import *

def extract(line,field):
    new_fiel = f"\"{field}\":"
    return line.split(new_fiel)[1].split()[0][:-1]


def extractTime(line):
    return int(line.split("\"time(micro)\":")[1].split()[0][:-1])

def getNode(line):
    return int(line.split()[0][-1])

# ================================================
#   General protocols
# ================================================


# Log line:
#   logTag  diffTag: tag  describers

TIME_DESCRIBER = "time(micro)"

class GeneralProtocol:

    def __init__(self, logger, name, logTag, tags, describers):
        self.logger = logger
        self.name = name
        self.logTag = logTag
        self.describers = describers
        self.tags = {tag:{} for tag in tags}
        self.diffTag = None
        self.lines = []
        self.ssv_node_num = None

        # run time variables
        self.length = None
        self.sorted_idx = None

    def setDiffTag(self,v):
        self.diffTag = v

    def __str__(self):
        return f"{self.name=},{self.logTag},{self.diffTag=},tags={self.tags.keys()},{self.describers=}"

    def add(self,line):


        # verify log tag
        if self.logTag not in line:
            self.logger.debug("quitting GeneralProtocol function because logTag not in line." + (str(self)) + f".{line=}")
            return
        
        
        # verify diff tag
        if self.diffTag != None:
            if self.diffTag not in line:
                self.logger.debug("quitting GeneralProtocol function because diffTag not in line." + (str(self)) + f".{line=}")
                return

        # verify describers
        describersValues = {}
        for describer in self.describers:
            try:
                describersValues[describer] = extract(line,describer)
            except:
                self.logger.debug(f"quitting GeneralProtocol function because {describer=} not in line." + (str(self)) + f".{line=}")
                return


        # verify tags
        found = False
        comment = line.split("$$$$$$")[1]
        tag_field = comment.split(self.diffTag)[1]

        for tag in self.tags:
            if tag in tag_field:
                if len(self.tags[tag]) != 0:
                    # self.logger.debug(f"jumping GeneralProtocol function because tried to add {tag=} but already has it {self.tags[tag]}." + (str(self)) + f".{comment=}.{line=}")
                    pass
                else:
                    # self.logger.debug(f"adding {tag=}")
                    self.tags[tag] = {desc: describersValues[desc] for desc in describersValues}
                    self.tags[tag]['comment'] = comment
                    found = True
                    break
        
        
        if found:
            node_num = int(line.split("ssv-node-")[1].split()[0])
            if self.ssv_node_num == None:
                self.ssv_node_num = node_num
            else:
                if self.ssv_node_num != node_num:
                    self.logger.debug(f"Warning: GeneralProtocol:Add node number different than self node number. {node_num=}, {line=}")
            self.lines.append(line)
        else:
            self.logger.debug(f"quitting GeneralProtocol function because found not tag in line." + (str(self)) + f".{line=}")
            return

        return
    
    def len(self):
        if self.length == None:
            self.analyse()
        return self.length

    def analyse(self):

        tags = list(self.tags.keys())
        times = []
        idx = []

        for i in range(len(tags)):
            if TIME_DESCRIBER in self.tags[tags[i]]:
                times += [int(self.tags[tags[i]][TIME_DESCRIBER])]
                idx.append(i)

        if idx == []:
            self.logger.debug(f"quitting GeneralProtocol:analyse function because there's no tag with time. {self.tags=}")
            return
    
        # sorted index
        idx = sorted(x for _, x in zip(times,idx))
    

        initial_time = int(self.tags[tags[idx[0]]][TIME_DESCRIBER])
        final_time = int(self.tags[tags[idx[-1]]][TIME_DESCRIBER])
        self.length = final_time - initial_time

        self.sorted_idx = idx
    
    def show(self):
        if self.sorted_idx == None:
            self.analyse()
        if self.sorted_idx == None:
            self.logger.debug(f"quitting GeneralProtocol:show function because sorted_idx continued None after self.analyse().")
            return
        
        tags = list(self.tags.keys())

        text = ""

        last_time = None
        last_name = None
        for i in range(len(self.sorted_idx)):
            if i == 0:
                text += f"Show {BLUE}{self.name}{END} - {self.diffTag} ({PURPLE}{self.ssv_node_num}, {self.tags[list(self.tags.keys()[0])]}{END}):\n"
                last_time = int(self.tags[tags[self.sorted_idx[i]]][TIME_DESCRIBER])
                last_name = tags[self.sorted_idx[i]]
                text += f"\tstart at {last_name}: {last_time}\n"
            else:
                curr_time = int(self.tags[tags[self.sorted_idx[i]]][TIME_DESCRIBER])
                curr_name = tags[self.sorted_idx[i]]
                text += f"\t\tthen {last_name} - {curr_name}: {curr_time} (d:{curr_time-last_time})\n"
                last_name = curr_name
                last_time = curr_time
        
        text += f"\tTotal time: {YELLOW}{self.length}{END}\n"

        self.logger.debug(text)
    
    def showSummary(self):
        describers_values_aux = self.tags[list(self.tags.keys())[0]]
        describers_values = {key: describers_values_aux[key] for key in describers_values_aux if key != "comment"}
        title = f"{LIGHT_BLUE}ssv-node-{END}{LIGHT_CYAN}{self.ssv_node_num}{END} {PURPLE}{self.logTag}{END} {LIGHT_WHITE}{describers_values}{END}"
        self.logger.debug(title)

        tags = list(self.tags.keys())
        last_time = int(self.tags[tags[self.sorted_idx[0]]][TIME_DESCRIBER])
        for i in range(len(self.sorted_idx)):
            new_time = int(self.tags[tags[self.sorted_idx[i]]][TIME_DESCRIBER])
            comment = self.tags[tags[self.sorted_idx[i]]]['comment']
            self.logger.debug(f"\t{comment}: {LIGHT_CYAN}{new_time - last_time}{END}")
            last_time = new_time
        
        total_time = int(self.tags[tags[self.sorted_idx[-1]]][TIME_DESCRIBER]) - int(self.tags[tags[self.sorted_idx[0]]][TIME_DESCRIBER])
        self.logger.debug(f"\t {LIGHT_WHITE}Total time{END}: {YELLOW}{total_time}{END}")

    def hasDecidedTime(self):
        # if self.ssv_node_num != 1:
        #     return None
        if self.logTag != 'UponABAFinish':
            return None
        
        for tag in self.tags:
            if 'comment' not in self.tags[tag]:
                continue
            if 'return decided, total time: ' in self.tags[tag]['comment']:
                txt = self.tags[tag]['comment']
                return int(txt.split("return decided, total time: ")[1])

        return None
    
    def getMinTime(self):
        times = [int(self.tags[tag][TIME_DESCRIBER]) for tag in self.tags if TIME_DESCRIBER in self.tags[tag]]
        return min(times)
    
    def getMaxTime(self):
        times = [int(self.tags[tag][TIME_DESCRIBER]) for tag in self.tags if TIME_DESCRIBER in self.tags[tag]]
        return max(times)


class GeneralProtocolController:

    def __init__(self,logger):
        self.logger = logger
        self.protocols = []
        self.tags = ['UponBaseRunnerDecide','UponStart','UponProposal','UponPrepare','UponCommit','UponRoundChange','UponRoundTimeout']
        self.protocols_dict = {tag:{} for tag in self.tags}

    def lenMessages(self):
        return len(self.protocols)    

    def getAll(self):
        return self.protocols

    def add(self,line):
        line_tag = None
        for tag in self.tags:
            if tag in line:
                line_tag = tag
                break
        
        if line_tag == None:
            self.logger.debug(f"quitting GeneralProtocolController:add because there's no tag in line. {line=}.{self.tags=}")
            return

        diffTag = line.split(f"{line_tag} ")[1].split(":")[0]

        if diffTag in self.protocols_dict[line_tag]:
            self.protocols_dict[line_tag][diffTag].add(line)
        else:
            p = self.createProtocol(line_tag,diffTag)
            self.protocols += [p]
            p.add(line)
            self.protocols_dict[line_tag][diffTag] = p

    def analyse(self):
        for p in self.protocols:
            p.analyse()
        
    def show(self):
        for p in self.protocols:
            p.show()
        
    def createProtocol(self,tag,diffTag):
        p = None
        if tag == 'UponStart':
            p = self.createStart()
        elif tag == 'UponBaseRunnerDecide':
            p = self.createBaseRunner()
        elif tag == 'UponProposal':
            p = self.createProposal()
        elif tag == 'UponPrepare':
            p = self.createPrepare()
        elif tag == 'UponCommit':
            p = self.createCommit()
        elif tag == 'UponRoundChange':
            p = self.createRoundChange()
        elif tag == 'UponRoundTimeout':
            p = self.createTimeout()
        p.setDiffTag(diffTag)
        return p

    def createBaseRunner(self):
        return GeneralProtocol(self.logger,'BaseRunner','UponBaseRunnerDecide',
                        ['start new instance'],
                         ['time(micro)'])

    def createStart(self):
        return GeneralProtocol(self.logger,'Start','UponStart',
                        ['preStartOnce',
                         'start qbft instance',
                         'create proposal',
                         'broadcast start',
                         'finish'],
                         ['time(micro)'])

    def createProposal(self):
        return GeneralProtocol(self.logger,'Proposal','UponProposal',
                        ['start',
                         'addFirstMsg',
                         'create prepare msg',
                         'broadcast start',
                         'broadcast finish',
                         'finish'],
                         ['time(micro)',
                          'round',
                          'sender'])

    def createPrepare(self):
        return GeneralProtocol(self.logger,'Prepare','UponPrepare',
                        ['start',
                         'get proposal data',
                         'add signed msg',
                         'check if havent quorum',
                         'check if sent commit for height and round',
                         'create commit msg',
                         'broadcast start',
                         'broadcast finish',
                         'finish'],
                         ['time(micro)',
                          'round',
                          'sender'])
    
    def createCommit(self):
        return GeneralProtocol(self.logger,'Commit','UponCommit',
                        ['start',
                         'add signed message',
                         'calculate commit',
                         'got quorum. Getting commit data',
                         'aggregate commit',
                         'return decided, total time:',
                         'finish no quorum'],
                         ['time(micro)',
                          'round',
                          'sender'])
    
    def createRoundChange(self):
        return GeneralProtocol(self.logger,'RoundChange','UponRoundChange',
                        ['start',
                         'add signed message',
                         'check proposal justification',
                         'create proposal',
                         'broadcast start',
                         'broadcast finish',
                         'round change partial quorum',
                         'finish'],
                         ['time(micro)',
                          'round',
                          'sender'])

    def createTimeout(self):
        return GeneralProtocol(self.logger,'Timeout','UponRoundTimeout',
                        ['start',
                         'create round change msg',
                         'broadcast start',
                         'broadcast finish',
                         'finish'],
                         ['time(micro)',
                          'round'])









# ================================================
#   Execution
# ================================================

class Execution:

    def __init__(self,logger,publicKey = None, role = None, height = None):
        self.logger = logger
        self.publicKey = publicKey
        self.role = role
        self.height = height
        self.lines = []

        self.subprotocolControllers = {}
        self.ctrl = None

    def getId(self):
        return f"({self.publicKey}, {self.role}, {self.height})"

    def add(self,line):
        self.lines.append(line)
    
    def show(self):
        self.logger.debug(f"{self.publicKey=},{self.role=},{self.height=}")
        for line in self.lines:
            self.logger.debug(line)
    
    def showSummary(self):
        for prot in self.ctrl.getAll():
            prot.showSummary()

    def showTimeMetrics(self):
        prots = self.ctrl.getAll()
        min_time = prots[0].getMinTime()
        max_time = prots[0].getMaxTime()

        for prot in prots:
            min_time = min(min_time,prot.getMinTime())
            max_time = max(max_time,prot.getMaxTime())
        
        len_time = max_time - min_time
        self.logger.debug(f"{min_time=}, {max_time=}, {LIGHT_CYAN}{len_time=}{END}")

        times = []
        for prot in prots:
            v = prot.hasDecidedTime()
            if v != None:
                times += [v]
        
        for latency_time in times:
            self.logger.debug(f"Latency time: {latency_time}")
        
        self.logger.debug(f"Mean latency time: {sum(times)/len(times)}")
        
        self.logger.debug(f"Number of decided: {len(times)}. All execution latency: {len_time/len(times)}. All execution throughput {len(times)/len_time}")

    def getMinTime(self):
        prots = self.ctrl.getAll()
        min_time = prots[0].getMinTime()

        for prot in prots:
            min_time = min(min_time,prot.getMinTime())

        return min_time

    def getMaxTime(self):
        prots = self.ctrl.getAll()
        max_time = prots[0].getMaxTime()

        for prot in prots:
            max_time = max(max_time,prot.getMaxTime())
            
        return max_time

    def getLenTime(self):
        prots = self.ctrl.getAll()
        min_time = prots[0].getMinTime()
        max_time = prots[0].getMaxTime()

        for prot in prots:
            min_time = min(min_time,prot.getMinTime())
            max_time = max(max_time,prot.getMaxTime())
        
        len_time = max_time - min_time
        return len_time

    def getTimes(self):
        prots = self.ctrl.getAll()
        times = []
        for prot in prots:
            v = prot.hasDecidedTime()
            if v != None:
                times += [v]
        return times


    def showController(self):
        self.ctrl.show()

    def analyse(self, verbose = True):

        self.ctrl = GeneralProtocolController(self.logger)

        for line in self.lines:
            self.ctrl.add(line)

        self.ctrl.analyse()
    
    def lenMessages(self):
        return self.ctrl.lenMessages()


    def showAnalysis(self):
        self.ctrl.show()

    def ganttMatplotlib(self,save = False):

        # build dictionary with tasks
        data = {
            'Task': [],
            'Round': [],
            'Sender': [],
            'Start': [],
            'Finish': [],
            'Duration': [],
            # 'Broadcast Start': [],
            # 'Broadcast Finish': [],
            # 'Aggregate Start': [],
            # 'Aggregate Finish': []
        }

        for subprotocol_execution in self.subprotocol_executions:
            # subprotocol_execution.showSingleton(self.start)
            if subprotocol_execution.start != None:
                data['Task'].append(subprotocol_execution.getName())
                data['Sender'].append(subprotocol_execution.sender)
                data['Round'].append(subprotocol_execution.round)
                data['Start'].append(subprotocol_execution.start)
                data['Finish'].append(subprotocol_execution.finish)
                data['Duration'].append(subprotocol_execution.len())
            # data['Broadcast Start'].append(subprotocol_execution.broadcastStart)
            # data['Broadcast Finish'].append(subprotocol_execution.broadcastFinish)
            # data['Aggregate Start'].append(subprotocol_execution.aggregateStart)
            # data['Aggregate Finish'].append(subprotocol_execution.aggregateFinish)
        
        print(data['Start'])
        print(data['Start'])
        for i in range(len(data['Task'])):
            print(f"{data['Task'][i]=},{data['Start'][i]=},{data['Duration'][i]=}")

        # get metrics for chart
        min_time = min(filter(None, data['Start'])) - 2000
        max_time = max(filter(None, data['Finish'])) + 5000
        print(min_time)
        print(max_time)


        task_names = list(set(data['Task']))
        ordered_task_names = []
        for task_name in data['Task']:
            if task_name not in ordered_task_names:
                ordered_task_names += [task_name]
        num_tasks = len(task_names)
        

        bg_color = "#34454e"

        # Create a colormap
        colormap = plt.cm.get_cmap('tab20c', num_tasks)
        

        # generate more info about tasks (duration, broadcast duration, aggregate duration)
        # duration = []
        # for i in range(len(data['Start'])):
        #     duration += [data['Finish'][i]-data['Start'][i]]
        # data['Duration'] = duration

        # bc_task = []
        # bc_start = []
        # bc_duration = []
        # for i in range(len(data['Broadcast Start'])):
        #     if data['Broadcast Start'][i] != None:
        #         bc_task += [data['Task'][i]]
        #         bc_start += [data['Broadcast Start'][i]]
        #         if data['Broadcast Finish'][i] != None:
        #             bc_duration += [data['Broadcast Finish'][i]-data['Broadcast Start'][i]]
        #         else:
        #             bc_duration += [max_time - data['Broadcast Start'][i]]


        # ag_task = []
        # ag_start = []
        # ag_duration = []
        # for i in range(len(data['Aggregate Start'])):
        #     if data['Aggregate Start'][i] != None:
        #         ag_task += [data['Task'][i]]
        #         ag_start += [data['Aggregate Start'][i]]
        #         if data['Aggregate Finish'][i] != None:
        #             ag_duration += [data['Aggregate Finish'][i]-data['Aggregate Start'][i]]
        #         else:
        #             ag_duration += [max_time - data['Aggregate Start'][i]]




        # Create a figure and an axis object
        fig, ax = plt.subplots(facecolor=bg_color)

        # Set the x-axis limits and labels
        ax.set_xlim(min_time, max_time)
        ax.set_xlabel('Time (microseconds)')

        # set facecolor and grid
        ax.set_facecolor(bg_color)
        ax.xaxis.grid(True)

        # set other colors
        ax.spines['bottom'].set_color('white')
        ax.spines['top'].set_color(bg_color)
        ax.spines['left'].set_color(bg_color)
        ax.spines['right'].set_color(bg_color)
        ax.xaxis.label.set_color('white')
        ax.yaxis.label.set_color('white')
        ax.tick_params(axis='x', colors='white')
        ax.tick_params(axis='y', colors='white')


        # create color map for tasks
        cmap = plt.colormaps["tab20c"]
        color_map = {}
        idx = 0
        for label in task_names:
            color_map[label] = cmap([idx])[0]
            idx = (idx+4)%20
        colors = [color_map[task_name] for task_name in data['Task']]

        ax.barh(data['Task'],data['Duration'],left = data['Start'],color = colors,edgecolor = colors,linewidth=1) # ['black']*len(data['Task'])


        # add legend
        legend_elements = [Patch(facecolor=color_map[i], label=i)  for i in color_map]
        plt.legend(handles=legend_elements)
        ax.set_title("Gantt chart",color='white')

        # create text with duration
        # for i in range(len(data['Task'])):
        #     y_loc = 0
        #     for task_name in ordered_task_names:
        #         if task_name == data['Task'][i]:
        #             break
        #         y_loc += 1
        #     ax.text(data['Finish'][i] + 100, y_loc,str(data['Duration'][i]),va='center',alpha = 0.9,color='white')


        # create color map for broadcast and aggregate bars and plot it
        color_map = {}
        idx = 2
        for label in task_names:
            color_map[label] = cmap([idx])[0]
            idx = (idx+4)%20
        # colors = [color_map[task_name] for task_name in bc_task]
        # ax.barh(bc_task, bc_duration, left=bc_start, color=colors, alpha=1)
        # colors = [color_map[task_name] for task_name in ag_task]
        # ax.barh(ag_task, ag_duration, left=ag_start, color=colors, alpha=1)

        # Display the plot
        if save:
            plt.savefig(f"images/{self.publicKey}_{self.role}_{self.height}_ganttchart.png")
        plt.show()



class ExecutionController:

    def __init__(self,logger):
        self.logger = logger
        self.executions = []
        self.publicKeyMap = {}

    
    def addExecution(self,publicKey = None, role = None, height = None):
        if publicKey not in self.publicKeyMap:
            self.publicKeyMap[publicKey] = {}
        if role not in self.publicKeyMap[publicKey]:
            self.publicKeyMap[publicKey][role] = {}
        
        if height not in self.publicKeyMap[publicKey][role]:
            execution = Execution(logger = self.logger, publicKey = publicKey, role = role, height = height)
            self.publicKeyMap[publicKey][role][height] = execution
            self.executions.append(execution)
            return execution
        else:
            return self.publicKeyMap[publicKey][role][height]

    def analyseExecutions(self):
        for ex in self.executions:
            ex.analyse()


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
    
    def get(self,pbkey,role,height):
        for e in self.executions:
            if e.publicKey == pbkey and e.role == role and e.height == height:
                return e
        return None

    def getAt(self,index):
        return self.executions[index]

    def getAll(self):
        return self.executions

    def len(self):
        return len(self.executions)

    def __len__(self):
        return self.len()

    def analyseAll(self):
        for e in self.executions:
            e.analyse()

    def describe(self):
        count = {}
        for idx,e in enumerate(self.executions):
            v = e.lenMessages()
            self.logger.debug(f"{e.getId()}: index {idx} :{v} messages")
            if v in count:
                count[v]+=1
            else:
                count[v]=1
        
        self.logger.debug("="*60)
        
        keys = count.keys()

        keys = sorted(keys,reverse=True)

        for key in keys:
            self.logger.debug(f"{count[key]} exeutions with {key} messages.")
    
    def showOverallTimeMetrics(self):

        # execution: {min_time, max_time, len_time, times, avg_time}
        exec_dicts = {}
        for e in self.executions:
            e_id = e.getId()
            exec_dicts[e_id] = {
                'min_time':e.getMinTime(),
                'max_time':e.getMaxTime(),
                'len_time':e.getLenTime(),
                'times':e.getTimes()
            }
            times = exec_dicts[e_id]['times']
            if times != []:
                exec_dicts[e_id]['mean_latency_time'] = sum(times)/len(times)
                exec_dicts[e_id]['mean_throughput_time'] = len(times)/sum(times)
                exec_dicts[e_id]['all_latency'] = exec_dicts[e_id]['len_time']/len(times)
                exec_dicts[e_id]['all_throughput'] = len(times)/(exec_dicts[e_id]['len_time']/1e6)

        min_all = min([e['min_time'] for e in exec_dicts.values()])
        max_all = max([e['max_time'] for e in exec_dicts.values()])
        num_decided = sum([len(e['times']) for e in exec_dicts.values()])

        self.logger.debug(f"Min time: {min_all}. Max time: {max_all}. Total duration: {max_all-min_all}")
        self.logger.debug(f"All time mean Latency: {(max_all-min_all)/num_decided}")
        self.logger.debug(f"All time mean Throughput: {num_decided/((max_all-min_all)/1e6)}")

        times = []
        for e in exec_dicts.values():
            times += e['times']
        

        self.logger.debug(f"Latency times ({len(times)}):")
        for e_id in exec_dicts:
            if len(exec_dicts[e_id]['times']) != 0:
                self.logger.debug(f"{e_id}: {exec_dicts[e_id]['times']}")
        self.logger.debug(f"Mean latency time: {sum(times)/len(times)}")


class Analyzer:

    def __init__(self,logger,filename = "out.txt"):

        self.logger = logger

        self.lines = []
        with open(filename,"r") as f:
            self.lines = f.readlines()
        
    
    def show(self):
        for line in self.lines:
            print(line)
    
    def len(self):
        return len(self.lines)
    
    def getExecutionController(self):
        

        exec_ctrl = ExecutionController(self.logger)

        # get instance parameters
        def getExecutionParams(line):
            ans = {}
            terms = ["\"publicKey\":","\"role\":","\"height\":"]
            for term in terms:
                if term in line:
                    v = line.split(term)[1].split()[0].split(",")[0]
                    ans[term] = v
            ans["\"publicKey\":"] = ans["\"publicKey\":"][1:-1]
            ans["\"role\":"] = ans["\"role\":"][1:-1]
            return ans

        # for each line, add the execution and add the line
        for line in self.lines:
            try:
                execParams = getExecutionParams(line)
                execution = exec_ctrl.addExecution(execParams["\"publicKey\":"],execParams["\"role\":"],execParams["\"height\":"])
                execution.add(line)
            except Exception as e:
                print(f"Exception while adding execution: {e=}. {line=}")

        self.logger.debug(f'Has created {len(exec_ctrl)} executions')

        return exec_ctrl

    def filter(self,value):
        i = 0
        ans = []
        while i < len(self.lines):
            if value in self.lines[i]:
                ans += [self.lines[i]]
            i = i + 1
        self.lines = ans



def executionAnalyser(file,node):

    logger = Logger("analysis.log",verbose_mode=1)

    a = Analyzer(logger,filename = file)
    if node != None:
        a.filter(f'ssv-node-{node}')
    a.filter('$$$$$$')

    exec_ctrl = a.getExecutionController()

    exec_ctrl.analyseAll()
    
    exec_ctrl.describe()

    while True:
        cmd = int(input(
        "Enter a command:\n\
\t1: Show Execution\n\
\t2: Iterate\n\
\t3: Show execution summary\n\
\t4: Show execution time metrics\n\
\t5: Show overall execution time\n\
\t-1: Quit\n\
> \
"))
        if cmd == 1:
            id_tuple = input("public key, role, height (or index):")
            e = None
            if ", " in id_tuple:
                pbkey = id_tuple.split(", ")[0]
                role = id_tuple.split(", ")[1]
                height = id_tuple.split(", ")[2]
                e = exec_ctrl.get(pbkey,role,height)
            else:
                e = exec_ctrl.getAt(int(id_tuple))
            if e == None:
                print("Not found")
                continue
            e.show()
        elif cmd == 2:


            idx = 0
            while True:
                cmd = int(input("0 to continue; 1 to stop;"))
                if cmd == 0:
                    execution = exec_ctrl.getAt(idx)
                    execution.show()
                    # execution.analyse()
                    execution.showAnalysis()
                    # execution.ganttPlotly()
                    # execution.ganttMatplotlib(True)
                    # execution.barChart(True)
                    # execution.pieChart(True)
                    # execution.describe()
                    idx += 1
                else:
                    break
        elif cmd == 3:
            id_tuple = input("public key, role, height(or index):")
            e = None
            if ", " in id_tuple:
                pbkey = id_tuple.split(", ")[0]
                role = id_tuple.split(", ")[1]
                height = id_tuple.split(", ")[2]
                e = exec_ctrl.get(pbkey,role,height)
            else:
                e = exec_ctrl.getAt(int(id_tuple))
            if e == None:
                print("Not found")
                continue
            e.showSummary()
        elif cmd == 4:
            id_tuple = input("public key, role, height(or index):")
            e = None
            if ", " in id_tuple:
                pbkey = id_tuple.split(", ")[0]
                role = id_tuple.split(", ")[1]
                height = id_tuple.split(", ")[2]
                e = exec_ctrl.get(pbkey,role,height)
            else:
                e = exec_ctrl.getAt(int(id_tuple))
            if e == None:
                print("Not found")
                continue
            e.showTimeMetrics()
        elif cmd == 5:
            exec_ctrl.showOverallTimeMetrics()
        elif cmd == -1:
            break

    # a.show()
    print(f"Analyser len: {a.len()}")


def overallStatistics(file,node):
    a = Analyzer(filename = file)
    if node != None:
        a.filter(f'ssv-node-{node}')
    a.filter('$$$$$$')

    exec_ctrl = a.getExecutionController()

    print(f"Number of executions: {len(exec_ctrl.getAll())}")

    dict_values = exec_ctrl.getStatisticsData()
    mean_total_time, stddev, n_total_values, percentages = exec_ctrl.getMetricsOverStatisticsData(dict_values,plot=True,save=True)

    # print(dict_values)




parser = argparse.ArgumentParser()

subparsers = parser.add_subparsers(dest='command', required=True)

# 'overall' sub-command
overall_parser = subparsers.add_parser('overall')
overall_parser.add_argument('-f', '--file', type=str, required=True, help='file path')
overall_parser.add_argument('-n', '--node', type=int, help='node count')

# 'executions' sub-command
executions_parser = subparsers.add_parser('executions')
executions_parser.add_argument('-f', '--file', type=str, required=True, help='file path')
executions_parser.add_argument('-n', '--node', type=int, help='node count')



if __name__ == "__main__":

    args = parser.parse_args()

    if args.command == 'overall':
        print(f"Overall command with file {args.file} and node {args.node}")
        overallStatistics(args.file,args.node)


    elif args.command == 'executions':
        print(f"Executions command with file {args.file} and node {args.node}")
        executionAnalyser(args.file,args.node)