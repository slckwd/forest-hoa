import json

import dateutil.parser
from dateutil.relativedelta import relativedelta

import datetime as dt
import pandas as pd
import pandas_ods_reader as pods


class BankData:
    file_path = '../'
    load_file = 'load_file.csv'
    sheet_name = 'Sheet 1'
    output_file = 'output_file.csv'
    desc_col = 'desc_col'
    date_col = 'date_col'
    amnt_col = 'amnt_col'
    regex_col = 'regex_col'
    regex_cols = [ ]
    ndate_col = 'Normalized Date'
    summ_col = 'Summary Total'
    sum_col = 'Sum'
    total_col = 'Total'
    labels = { }
    stopdate = dt.date(2017, 1, 1)
    
    transactions = pd.DataFrame()
    summary = pd.DataFrame(columns=[ndate_col])
    
    def __init__(self):
        self.loadconfig()
        file = pods.read_ods(self.file_path + self.load_file, sheet=self.sheet_name)
        file[self.ndate_col] = file.apply(lambda row: self.convert(row[self.date_col]), axis=1)
        file[self.regex_col] = file.apply(lambda row: self.regexstring(row), axis=1)
        self.transactions = file
        
        date = dt.date(dt.datetime.today().year, dt.datetime.today().month, 1)
        while date >= self.stopdate:
            self.summary = self.summary.append({self.ndate_col: str(date)}, ignore_index=True)
            date = date + relativedelta(months=-1)
    
    def summarize(self):
        for entry in self.labels:
            payments = self.transactions.loc[self.transactions.loc[:, self.regex_col].str.contains(pat=entry['regex'], regex=True)].copy()
            grouped = payments.groupby([self.ndate_col])[self.amnt_col].sum().reset_index().rename(columns={self.amnt_col: entry['label']})
            self.summary = self.summary.merge(grouped, how='left', on=self.ndate_col).fillna(0)
        
        self.summary[self.summ_col] = self.summary.loc[:, self.summary.columns.difference([self.ndate_col, self.total_col])].sum(axis=1)
        self.summary[self.sum_col] = self.summary[self.total_col] - self.summary[self.summ_col]
        self.summary[self.sum_col] = self.summary[self.sum_col].round(2)
        self.summary.insert(len(self.summary.columns) - 3, '', '')
    
    def loadconfig(self):
        with open('../config.json') as json_file:
            config = json.load(json_file)
            self.file_path = config['filePath']
            self.load_file = config['fileName']
            self.output_file = config['outputFile']
            self.sheet_name = config['sheetName']
            self.desc_col = config['descriptionColumn']
            self.amnt_col = config['amountColumn']
            self.date_col = config['dateColumn']
            self.regex_cols = config['regexColumns']
            self.labels = config['labels']
            self.labels.append({ 'label': self.total_col, 'regex': '(?i:.)' })
            self.stopdate = dateutil.parser.parse(config['stopDate']).date()
    
    def convert(self, entry): 
        date = dateutil.parser.parse(entry)
        return str.format('{:04}-{:02}-01', date.year, date.month)
    
    def regexstring(self, entry):
        value = ''
        for column in self.regex_cols:
            value += str(entry[column]) + '_'
        return value.rstrip('_')
    
    def export(self):
        self.summary.to_csv(self.file_path + self.output_file, index=False)


data = BankData()

data.summarize()
data.export()
