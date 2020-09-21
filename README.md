Personal project for managing my HOA finances.

Reads an arbitrary csv of bank transactions and organizes them into configurable categories/people/entities on a monthly basis.

Reads in a configuration file in the format:
```json
{
	"input": {
		"path": "C:/Users/Midas/Document",
		"name": "transactions.csv"
	},
	"output": {
		"path": "C:/Users/Midas/Documents/",
		"name": "summary.csv"
	},
	"columns": {
		"description": 1,
		"debit": 2,
		"credit": 3,
		"date": 4,
		"fixedAmounts": {
			"Lawn": 20
		}
	},
	"labels": [
		{ "label": "Internet", "regex": "(?i:Starlink)" },
		{ "label": "Water", "regex": "(?i:City)" },
		{ "label": "Garbage", "regex": "(?i:Waste)" },
        { "label": "Electricity", "regex": "(?i:Power)" },
        { "label": "Paycheck", "regex": "(?i:Pay)" },
        { "label": "Lawn", "regex": "(?i:UNSEARCHABLE)" }
	],
	"stopDate": "2017-01-01"
}
```
- `input` and `output`: The files to be read from and written to.
  - `path`: The directory of the file.
  - `name`: The name of the file.
- `columns`: The columns in the input file correspoding to each piece of data.
  - `description`: The description of the transaction.
  - `debit`: The amount deducted.
  - `credit`: The amount added.
  - `date`: The date of the transaction.
  - `fixedAmounts`: For transactions with poorly searchable descriptions that can instead be identified by the amount of the transaction.
- `labels`: Human readable labels and regex to match against descriptions.
  - `label`: The column label for matching transactions.
  - `regex`: A regular expression to match against the description.
- `stopDate`: The earliest date to include in the output.