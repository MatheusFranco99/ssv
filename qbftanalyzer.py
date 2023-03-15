
import plotly.express as px
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from datetime import datetime, timedelta
from matplotlib.patches import Patch
import statistics

def extract(line,field):
    return line.split(field)[1].split()[0][:-1]


def extractTime(line):
    return int(line.split("\"time(micro)\":")[1].split()[0][:-1])

# ================================================
#   Subprotocol
# ================================================

class Subprotocol:
    def __init__(self,sender,round):
        self.sender = sender
        self.round = round
        self.start = None
        self.finish = None
        self.hasBroadcast = False
        self.broadcastStart = None
        self.broadcastFinish = None
        self.showedSingleton = False
        self.hasAggregate = False
        self.aggregateStart = None
        self.aggregateFinish = None
    
    def getName(self):
        return "BaseSubprotocol"
    
    def process(self,line):
        if "broadcast start" in line:
            self.hasBroadcast = True
            self.broadcastStart = extractTime(line)
        elif "broadcast finish" in line:
            self.broadcastFinish = extractTime(line)
        elif f"{self.getName()} start" in line:
            self.start = extractTime(line)
        elif f"{self.getName()} return" in line:
            self.finish = extractTime(line)
    
    def len(self):
        if self.finish != None:
            return self.finish - self.start
        if self.broadcastFinish != None:
            return self.broadcastFinish - self.start
        if self.broadcastStart != None:
            return self.broadcastStart - self.start
        return 0
    
    def show(self):
        print(f"{self.getName()} {self.sender=} {self.round=}")
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
            print(f"{self.getName()} {self.sender=} {self.round=}")
            print(f"\t{self.start=}, time: {self.start - t0}")
            t = self.start
            if self.hasBroadcast:
                print(f"\t{self.broadcastStart=}, time: {self.broadcastStart - t0}, delta:{self.broadcastStart-t}")
                t = self.broadcastStart
            if self.broadcastFinish != None:    
                print(f"\t{self.broadcastFinish=}, time: {self.broadcastFinish - t0}, delta:{self.broadcastFinish-t}")
                t = self.broadcastFinish
            if self.finish != None:
                print(f"\t{self.finish=}, time: {self.finish - t0}, delta:{self.finish-t}")
                print(f"total: {self.len()}")

    def reset(self):
        self.showedSingleton = False


class UponChangeRound(Subprotocol):

    def __init__(self,sender,round):
        super().__init__(sender,round)
    
    def getName(self):
        return "UponRoundChange"


class UponRoundTimeout(Subprotocol):

    def __init__(self,sender,round):
        super().__init__(sender,round)
    
    def getName(self):
        return "UponRoundTimeout"


class UponCommit(Subprotocol):

    def __init__(self,sender,round):
        super().__init__(sender,round)

    def getName(self):
        return "UponCommit"
    
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
        if self.finish != None:
            return self.finish - self.start
        if self.aggregateFinish != None:
            return self.aggregateFinish - self.start
        if self.aggregateStart != None:
            return self.aggregateStart - self.start
        return 0
    
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
                t = self.aggregateStart
            if self.aggregateFinish != None:    
                print(f"\t{self.aggregateFinish=}, time: {self.aggregateFinish - t0}, delta:{self.aggregateFinish-t}")
                t = self.aggregateFinish
            if self.finish != None:
                print(f"\t{self.finish=}, time: {self.finish - t0}, delta:{self.finish-t}")
                print(f"total: {self.len()}")


class UponPrepare(Subprotocol):

    def __init__(self,sender,round):
        super().__init__(sender,round)
    
    def getName(self):
        return "UponPrepare"


class UponProposal(Subprotocol):

    def __init__(self,sender,round):
        super().__init__(sender,round)
    
    def getName(self):
        return "UponProposal"



# ================================================
#   SubprotocolController
# ================================================

class SubprotocolController:

    def __init__(self,creation_function):
        self.entries = {}
        self.creation_function = creation_function
    
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
            up = self.creation_function(sender,round)
            self.entries[sender][round] = up
            return up
    
    def getTotalTime(self):
        ans = 0
        for sender in self.entries:
            for round in self.entries[sender]:
                ans += self.entries[sender][round].len()
        return ans
    
    def reset(self):
        for sender in self.entries:
            for round in self.entries[sender]:
                self.entries[sender][round].reset()


class UponChangeRoundController(SubprotocolController):
    def __init__(self):
        super().__init__(lambda sender,round: UponChangeRound(sender,round))


class UponRoundTimeoutController(SubprotocolController):
    def __init__(self):
        super().__init__(lambda sender,round: UponRoundTimeout(sender,round))
    
    def add(self,round):
        sender = 0
        if sender not in self.entries:
            self.entries[sender] = {}
        
        if round in self.entries[sender]:
            return self.entries[sender][round]
        else:
            up = self.creation_function(sender,round)
            self.entries[sender][round] = up
            return up
    

    def add(self,round):
        sender = 0
        if sender not in self.entries:
            self.entries[sender] = {}
        
        if round in self.entries[sender]:
            return self.entries[sender][round]
        else:
            up = self.creation_function(sender,round)
            self.entries[sender][round] = up
            return up


class UponCommitController(SubprotocolController):
    def __init__(self):
        super().__init__(lambda sender,round: UponCommit(sender,round))


class UponPrepareController(SubprotocolController):
    def __init__(self):
        super().__init__(lambda sender,round: UponPrepare(sender,round))


class UponProposalController(SubprotocolController):
    def __init__(self):
        super().__init__(lambda sender,round: UponProposal(sender,round))


# ================================================
#   Execution
# ================================================

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

    def analyse(self, verbose = True):
        if verbose:
            self.show()
            print("="*90)

        self.uponProposalController = UponProposalController()
        self.uponPrepareController = UponPrepareController()
        self.uponCommitController = UponCommitController()
        self.uponChangeRoundController = UponChangeRoundController()
        self.uponRoundTimeoutController = UponRoundTimeoutController()

        self.subprotocol_executions = []

        for line in self.lines:
            if "UponProposal" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = self.uponProposalController.add(sender,round)
                up.process(line)
                if up not in self.subprotocol_executions:
                    self.subprotocol_executions += [up]
            if "UponPrepare" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = self.uponPrepareController.add(sender,round)
                up.process(line)
                if up not in self.subprotocol_executions:
                    self.subprotocol_executions += [up]
            if "UponCommit" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = self.uponCommitController.add(sender,round)
                up.process(line)
                if up not in self.subprotocol_executions:
                    self.subprotocol_executions += [up]
            if "UponRoundChange" in line:
                sender = extract(line,"\"sender\":")
                round = extract(line,"\"round\":")
                up = self.uponChangeRoundController.add(sender,round)
                up.process(line)
                if up not in self.subprotocol_executions:
                    self.subprotocol_executions += [up]
            if "UponRoundTimeout" in line:
                round = extract(line,"\"round\":")
                up = self.uponRoundTimeoutController.add(round)
                up.process(line)
                if up not in self.subprotocol_executions:
                    self.subprotocol_executions += [up]
            


        self.start = None
        self.finish = None
        for line in self.lines:
            if "starting QBFT instance" in line:
                self.start = extractTime(line)
            elif "Decided on value with commit" in line:
                self.finish = extractTime(line)

        if self.start == None:
            return     

        if verbose:
            for subprotocol_execution in self.subprotocol_executions:
                subprotocol_execution.showSingleton(self.start)

            self.reset()
        # print("="*30)

        # for line in self.lines:
        #     if "UponProposal" in line:
        #         sender = extract(line,"\"sender\":")
        #         round = extract(line,"\"round\":")
        #         up = self.uponProposalController.get(sender,round)
        #         if up != None:
        #             up.showSingleton(self.start)
        #     if "UponPrepare" in line:
        #         sender = extract(line,"\"sender\":")
        #         round = extract(line,"\"round\":")
        #         up = self.uponPrepareController.get(sender,round)
        #         if up != None:
        #             up.showSingleton(self.start)
        #     if "UponCommit" in line:
        #         sender = extract(line,"\"sender\":")
        #         round = extract(line,"\"round\":")
        #         up = self.uponCommitController.get(sender,round)
        #         if up != None:
        #             up.showSingleton(self.start)
        #     if "UponRoundChange" in line:
        #         sender = extract(line,"\"sender\":")
        #         round = extract(line,"\"round\":")
        #         up = self.uponChangeRoundController.get(sender,round)
        #         if up != None:
        #             up.showSingleton(self.start)
        #     if "UponRoundTimeout" in line:
        #         round = extract(line,"\"round\":")
        #         up = self.uponRoundTimeoutController.get(round)
        #         if up != None:
        #             up.showSingleton(self.start)
        

        if verbose:
            if self.finish != None:
                print(f"Total duration: {self.finish - self.start}")
    
    
    def totalTime(self):
        if self.start == None or self.finish == None:
            return None
        return self.finish - self.start

    def workTime(self):
        ans = 0
        for ctrl in [self.uponProposalController,
                    self.uponPrepareController,
                    self.uponCommitController,
                    self.uponChangeRoundController,
                    self.uponRoundTimeoutController]:
            ans += ctrl.getTotalTime()
        return ans

    def proposalTime(self):
        return self.uponProposalController.getTotalTime()

    def prepareTime(self):
        return self.uponPrepareController.getTotalTime()

    def commitTime(self):
        return self.uponCommitController.getTotalTime()

    def roundChangeTime(self):
        return self.uponChangeRoundController.getTotalTime()

    def timeoutTime(self):
        return self.uponRoundTimeoutController.getTotalTime()

    def reset(self):

        for ctrl in [self.uponProposalController,
                        self.uponPrepareController,
                        self.uponCommitController,
                        self.uponChangeRoundController,
                        self.uponRoundTimeoutController]:
            ctrl.reset()

    def ganttMatplotlib(self,save = False):

        # build dictionary with tasks
        data = {
            'Task': [],
            'Round': [],
            'Sender': [],
            'Start': [],
            'Finish': [],
            'Broadcast Start': [],
            'Broadcast Finish': [],
            'Aggregate Start': [],
            'Aggregate Finish': []
        }

        for subprotocol_execution in self.subprotocol_executions:
            # subprotocol_execution.showSingleton(self.start)
            data['Task'].append(subprotocol_execution.getName())
            data['Sender'].append(subprotocol_execution.sender)
            data['Round'].append(subprotocol_execution.round)
            data['Start'].append(subprotocol_execution.start)
            data['Finish'].append(subprotocol_execution.finish)
            data['Broadcast Start'].append(subprotocol_execution.broadcastStart)
            data['Broadcast Finish'].append(subprotocol_execution.broadcastFinish)
            data['Aggregate Start'].append(subprotocol_execution.aggregateStart)
            data['Aggregate Finish'].append(subprotocol_execution.aggregateFinish)
        

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
        duration = []
        for i in range(len(data['Start'])):
            duration += [data['Finish'][i]-data['Start'][i]]
        data['Duration'] = duration

        bc_task = []
        bc_start = []
        bc_duration = []
        for i in range(len(data['Broadcast Start'])):
            if data['Broadcast Start'][i] != None:
                bc_task += [data['Task'][i]]
                bc_start += [data['Broadcast Start'][i]]
                if data['Broadcast Finish'][i] != None:
                    bc_duration += [data['Broadcast Finish'][i]-data['Broadcast Start'][i]]
                else:
                    bc_duration += [max_time - data['Broadcast Start'][i]]


        ag_task = []
        ag_start = []
        ag_duration = []
        for i in range(len(data['Aggregate Start'])):
            if data['Aggregate Start'][i] != None:
                ag_task += [data['Task'][i]]
                ag_start += [data['Aggregate Start'][i]]
                if data['Aggregate Finish'][i] != None:
                    ag_duration += [data['Aggregate Finish'][i]-data['Aggregate Start'][i]]
                else:
                    ag_duration += [max_time - data['Aggregate Start'][i]]




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
        for i in range(len(data['Task'])):
            y_loc = 0
            for task_name in ordered_task_names:
                if task_name == data['Task'][i]:
                    break
                y_loc += 1
            ax.text(data['Finish'][i] + 100, y_loc,str(data['Duration'][i]),va='center',alpha = 0.9,color='white')


        # create color map for broadcast and aggregate bars and plot it
        color_map = {}
        idx = 2
        for label in task_names:
            color_map[label] = cmap([idx])[0]
            idx = (idx+4)%20
        colors = [color_map[task_name] for task_name in bc_task]
        ax.barh(bc_task, bc_duration, left=bc_start, color=colors, alpha=1)
        colors = [color_map[task_name] for task_name in ag_task]
        ax.barh(ag_task, ag_duration, left=ag_start, color=colors, alpha=1)

        # Display the plot
        if save:
            plt.savefig(f"images/{self.publicKey}_{self.role}_{self.height}_ganttchart.png")
        plt.show()



    def ganttPlotly(self):

        data = {
            'Task': [],
            'Round': [],
            'Sender': [],
            'Start': [],
            'Finish': [],
            'Broadcast Start': [],
            'Broadcast Finish': [],
            'Aggregate Start': [],
            'Aggregate Finish': []
        }

        for subprotocol_execution in self.subprotocol_executions:
            # subprotocol_execution.showSingleton(self.start)
            data['Task'].append(subprotocol_execution.getName())
            data['Sender'].append(subprotocol_execution.sender)
            data['Round'].append(subprotocol_execution.round)
            data['Start'].append(subprotocol_execution.start)
            data['Finish'].append(subprotocol_execution.finish)
            data['Broadcast Start'].append(subprotocol_execution.broadcastStart)
            data['Broadcast Finish'].append(subprotocol_execution.broadcastFinish)
            data['Aggregate Start'].append(subprotocol_execution.aggregateStart)
            data['Aggregate Finish'].append(subprotocol_execution.aggregateFinish)

        df = pd.DataFrame(data)
        print(df)

        # Convert the timestamps from Unix microseconds to datetime format
        df['Start'] = pd.to_datetime(df['Start'], unit='us')
        df['Finish'] = pd.to_datetime(df['Finish'], unit='us')
        df['Broadcast Start'] = pd.to_datetime(df['Broadcast Start'], unit='us')
        df['Broadcast Finish'] = pd.to_datetime(df['Broadcast Finish'], unit='us')
        df['Aggregate Start'] = pd.to_datetime(df['Aggregate Start'], unit='us')
        df['Aggregate Finish'] = pd.to_datetime(df['Aggregate Finish'], unit='us')

        # Create a Gantt chart using plotly
        fig = px.timeline(df, x_start='Start', x_end='Finish', y='Task', color='Sender')
        fig.update_yaxes(autorange="reversed")
        fig.show()

    def barChart(self,save = False):

        # x = ["proposal","prepare","commite","round_change","timeout"]
        # y = [ctrl.getTotalTime() for ctrl in [self.uponProposalController,
        #                                     self.uponPrepareController,
        #                                     self.uponCommitController,
        #                                     self.uponChangeRoundController,
        #                                     self.uponRoundTimeoutController]]

        # arrange values as {name1:[values],name2:[values],etc}
        labels = []
        values_map = {}
        for subprotocol_execution in self.subprotocol_executions:
            label = subprotocol_execution.getName()
            if label not in labels:
                labels.append(label)
                values_map[label] = [subprotocol_execution.len()]
            else:
                values_map[label] += [subprotocol_execution.len()]
        
        # normalize the length of each list of values
        max_len_values = 0
        for label in values_map:
            max_len_values = max(max_len_values,len(values_map[label]))
        
        for label in values_map:
            if max_len_values > len(values_map[label]):
                values_map[label] += [0] * (max_len_values - len(values_map[label]))

        
        # create figure
        fig, ax = plt.subplots()

        # plot bar
        bar_width = 0.8
        x_positions = np.arange(len(labels))
        bottom =[0] * len(labels)

        for i in range(max_len_values):
            values = [values_map[label][i] for label in values_map]
            ax.bar(x_positions, values, bottom=bottom, width=bar_width)
            bottom = [bottom[idx] + values[idx] for idx in range(len(bottom))]
        
        # put text with values in bars
        for bar in ax.patches:
            height = bar.get_height()
            width = bar.get_width()
            x = bar.get_x()
            y = bar.get_y()
            label_text = height
            label_x = x + width / 2
            label_y = y + height / 2
            ax.text(label_x, label_y, label_text, ha='center',    
                    va='center')

        # add description to chart
        ax.set_xticks(x_positions)
        x_labels = list(values_map.keys())
        ax.set_xticklabels(x_labels)
        ax.set_title('Distribution of Consumed Time per Subprotocol')
        ax.set_xlabel('Subprotocol')
        ax.set_ylabel('Time (microseconds)')

        if save:
            plt.savefig(f"images/{self.publicKey}_{self.role}_{self.height}_barchart.png")
        plt.show()


    
    def pieChart(self,save = False):

        # x = ["proposal","prepare","commite","round_change","timeout"]
        # y = [ctrl.getTotalTime() for ctrl in [self.uponProposalController,
        #                                     self.uponPrepareController,
        #                                     self.uponCommitController,
        #                                     self.uponChangeRoundController,
        #                                     self.uponRoundTimeoutController]]
    

        # x = []
        # y = []
        # for subprotocol_execution in self.subprotocol_executions:
        #     x += [subprotocol_execution.getName()]
        #     y += [subprotocol_execution.len()]
        
        # # normalize
        # total_y = sum(y)
        # y = [value_y / total_y for value_y in y]

        # # sort
        # y = [y for _,y in sorted(zip(x,y))]
        # x = sorted(x)

        # # colors
        # colors_values = ['red','blue','green','yellow','purple']
        # colors_map = {}
        # x_set = list(set(x))
        # i = 0
        # for value in x_set:
        #     colors_map[value] = colors_values[i]
        #     i += 1
        # colors = []
        # for i in range(len(x)):
        #     colors += [colors_map[x[i]]]

        

        # # create a pie chart
        # plt.pie(y, labels = x, autopct='%1.1f%%',colors=colors,shadow=True)
        # plt.title('Distribution of Consumed Time per Subprotocol')
        # plt.show()


        fig, ax = plt.subplots()
        size = 0.3

        # arrange values as {name1:[values],name2:[values],etc}
        labels = []
        y = {}
        for subprotocol_execution in self.subprotocol_executions:
            label = subprotocol_execution.getName()
            if label not in labels:
                labels.append(label)
                y[label] = [subprotocol_execution.len()]
            else:
                y[label] += [subprotocol_execution.len()]

        # outer colors and inner colors
        cmap = plt.colormaps["tab20c"]
        outer_colors_lst = []
        idx = 0
        for label in labels:
            outer_colors_lst += [idx]
            idx = (idx+4)%20
        outer_colors = cmap(outer_colors_lst)
        inner_colors_lst = []
        idx = 0
        for label in labels:
            v = outer_colors_lst[idx]
            for i in range(len(y[label])):
                inner_colors_lst += [v]
                v = (v+1)%20
            idx+=1
        inner_colors = cmap(inner_colors_lst)

        # total time by subprotocol
        ax.pie([sum(y[label]) for label in labels], autopct='%1.1f%%',labels=labels,normalize=True, radius=1, colors=outer_colors,wedgeprops={'linewidth': 3.0, 'edgecolor': 'white'})

        # times of each instance of subprotocol
        flatted_lst = []
        for lst in [y[label] for label in labels]:
            flatted_lst += lst
        ax.pie(flatted_lst,normalize=True, autopct='%1.1f%%', radius=1-size, colors=inner_colors,wedgeprops={'linewidth': 3.0, 'edgecolor': 'white'})

        ax.set(aspect="equal", title='Distribution of Consumed Time per Subprotocol')
        if save:
            plt.savefig(f"images/{self.publicKey}_{self.role}_{self.height}_piechart.png")
        plt.show()



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

    def getStatisticsData(self):
        print(f"Number of raw executions: {len(self.executions)}")
        ans = {'Total':[],
               'WorkTime':[],
               'ProposalTime':[],
               'PrepareTime':[],
               'CommitTime':[],
               'RoundChangeTime':[],
               'TimeoutTime':[]}
        for execution in self.executions:
            execution.analyse(verbose=False)
            total_time = execution.totalTime()
            work_time = execution.workTime()
            proposal_time = execution.proposalTime()
            prepare_time = execution.prepareTime()
            commit_time = execution.commitTime()
            round_change_time = execution.roundChangeTime()
            timeout_time = execution.timeoutTime()
            ans['Total'] += [total_time]
            ans['WorkTime'] += [work_time]
            ans['ProposalTime'] += [proposal_time]
            ans['PrepareTime'] += [prepare_time]
            ans['CommitTime'] += [commit_time]
            ans['RoundChangeTime'] += [round_change_time]
            ans['TimeoutTime'] += [timeout_time]
        return ans

    def getMetricsOverStatisticsData(self,statisticsData, plot = False, save = False):

        total_times = list(filter(None,statisticsData['Total']))
        mean_total_time = sum(total_times) / len(total_times)
        print(f"{mean_total_time=} over {len(total_times)} samples.")

        stddev = statistics.stdev(total_times)
        print(f"{stddev=} over {len(total_times)} samples.")

        percentages = {}
        for col in statisticsData:
            values = []
            for i in range(len(statisticsData['Total'])):
                if statisticsData['Total'][i] != None and statisticsData[col][i] != None:
                    values += [statisticsData[col][i]/statisticsData['Total'][i]]
            
            mean_v = sum(values) / len(values)
            stddev = statistics.stdev(values)
            print(f"for {col}, mean_percentage={mean_v} and {stddev=}, over {len(values)} samples.")

            percentages[col] = {}
            percentages[col]['mean'] = mean_v
            percentages[col]['stddev'] = stddev
            percentages[col]['max'] = max(values)
            percentages[col]['min'] = min(values)
            percentages[col]['n'] = len(values)
    
        if plot:
            # Total time - histogram
            fig = plt.figure()
            # Set the number of bins
            num_bins = 10

            # Create the histogram using plt.hist()
            n, bins, patches = plt.hist(total_times, bins=num_bins, density=True, alpha=0.75)

            # Set the title and axis labels
            plt.title('Execution Time Distribution')
            plt.xlabel('Execution Time (microseconds)')
            plt.ylabel('Frequency')
            plt.grid()

            # Show the plot

            if save:
                plt.savefig(f"images/overall_meantime_distribution.png")
            plt.show()

            # Percentages - bars

            means = [percentages[col]['mean'] for col in percentages if col != 'Total']
            stddevs = [percentages[col]['stddev'] for col in percentages if col != 'Total']
            n_groups = len([col for col in percentages if col != 'Total'])

            # create plot
            fig, ax = plt.subplots()
            index = np.arange(n_groups)
            bar_width = 0.35
            opacity = 0.8

            rects1 = plt.bar(index, means, bar_width,
            alpha=opacity,
            color='b',
            label='Mean percentage of occupation time')

            rects2 = plt.bar(index + bar_width, stddevs, bar_width,
            alpha=opacity,
            color='g',
            label='Stddev')

            plt.xlabel('Subtask')
            plt.ylabel('Percentage')
            plt.title('Mean percentage of occupation time')
            plt.xticks(index + bar_width, [col for col in percentages if col != 'Total'])
            plt.legend()

            plt.tight_layout()
            plt.grid()


            if save:
                plt.savefig(f"images/percentages_barchart.png")
            plt.show()



            # Percentages - candlesticks
            means = [percentages[col]['mean'] for col in percentages if col != 'Total']
            stddevs = [percentages[col]['stddev'] for col in percentages if col != 'Total']
            mins = [percentages[col]['min'] for col in percentages if col != 'Total']
            maxs = [percentages[col]['max'] for col in percentages if col != 'Total']
            n_groups = len([col for col in percentages if col != 'Total'])



            #create figure
            bg_color = "#e5e5e5"
            fig, ax = plt.subplots(facecolor=bg_color)

            ax.set_facecolor(bg_color)

            #define width of sticks elements
            width = .4
            width2 = .02


            cmap = plt.colormaps["tab20c"]
            color_idx = 0
            #plot
            for i in range(len([col for col in percentages if col != 'Total'])):

                color = cmap([color_idx])[0]
                color_idx = (color_idx+4)%20

                plt.bar(i,2 * stddevs[i],width,alpha=0.8,bottom=(means[i]-stddevs[i]),color=color,label=[col for col in percentages if col != 'Total'][i])
                plt.bar(i,maxs[i]-mins[i],width2,bottom=mins[i],color=color)

            #rotate x-axis tick labels
            plt.xticks(list(range(len([col for col in percentages if col != 'Total']))), [col for col in percentages if col != 'Total'])
            plt.grid(True)
            plt.legend()
            plt.xlabel('Subtask')
            plt.ylabel('Percentage')
            plt.title('Mean percentage of occupation time')

            #display chart
            if save:
                plt.savefig(f"images/percentages_sticks.png")
            plt.show()


        
        return mean_total_time,stddev,len(total_times),percentages





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
    
    def getExecutionController(self):
        

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


def executionAnalyser():
    a = Analyzer()
    a.filter('ssv-node-1')
    a.filter('$$$$$$')
    exec_ctrl = a.getExecutionController()
    idx = 0
    while True:
        cmd = int(input("0 to continue; 1 to stop;"))
        if cmd == 0:
            execution = exec_ctrl.getAt(idx)
            execution.analyse()
            # execution.ganttPlotly()
            execution.ganttMatplotlib(True)
            execution.barChart(True)
            execution.pieChart(True)
            idx += 1
        else:
            break

    # a.show()
    print(a.len())


def overallStatistics():
    a = Analyzer()
    a.filter('ssv-node-1')
    a.filter('$$$$$$')
    exec_ctrl = a.getExecutionController()

    dict_values = exec_ctrl.getStatisticsData()
    mean_total_time, stddev, n_total_values, percentages = exec_ctrl.getMetricsOverStatisticsData(dict_values,plot=True,save=True)

    # print(dict_values)


if __name__ == "__main__":
    # executionAnalyser()
    overallStatistics()