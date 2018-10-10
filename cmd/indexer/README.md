# Indexer

Indexer is a command to update the search indexes.

It will add new items and also trim off old ones.

```bash
$ go run cmd/indexer/main.go
```

## Cardiology

```
# Free, full-text articles...
loattrfree full text[Filter]

# From any of the followig journals...
AND (
"Am Heart J"[jour] OR
"Am J Cardiol"[jour] OR
"Arterioscler Thromb Vasc Biol"[jour] OR
"Atherosclerosis"[jour] OR
"Basic Res Cardiol"[jour] OR
"Cardiovasc Res"[jour] OR
"Chest"[jour] OR "Circulation"[jour] OR
"Circ Arrhythm Electrophysiol"[jour] OR
"Circ Cardiovasc Genet"[jour] OR
"Circ Cardiovasc Imaging"[jour] OR
"Circ Cardiovasc Qual Outcomes"[jour] OR
"Circ Cardiovasc Interv"[jour] OR
"Circ Heart Fail"[jour] OR
"Circ Res"[jour] OR
"ESC Heart Fail"[jour] OR
"Eur Heart J"[jour] OR
"Eur Heart J Cardiovasc Imaging"[jour] OR
"Eur Heart J Acute Cardiovasc Care"[jour] OR
"Eur Heart J Cardiovasc Pharmacother"[jour] OR
"Eur Heart J Qual Care Clin Outcomes"[jour] OR
"Eur J Heart Fail"[jour] OR
"Eur J Vasc Endovasc Surg"[jour] OR
"Europace"[jour] OR
"Heart"[jour] OR
"Heart Lung Circ"[jour] OR
"Heart Rhythm"[jour] OR
"JACC Cardiovasc Interv"[jour] OR
"JACC Cardiovasc Imaging"[jour] OR
"JACC Heart Fail"[jour] OR
"J Am Coll Cardiol"[jour] OR
"J Am Heart Assoc"[jour] OR
"J Am Soc Echocardiogr"[jour] OR
"J Card Fail"[jour] OR
"J Cardiovasc Electrophysiol"[jour] OR
"J Cardiovasc Magn Reson"[jour] OR
"J Heart Lung Transplant"[jour] OR
"J Hypertens"[jour] OR
"J Mol Cell Cardiol"[jour] OR
"J Thorac Cardiovasc Surg"[jour] OR
"J Vasc Surg"[jour] OR
"Nat Rev Cardiol"[jour] OR
"Prog Cardiovasc Dis"[jour] OR
"Resuscitation"[jour] OR
"Stroke"[jour]
)

# published between these dates - ie 12 month window...
AND ("2017/09/05"[PDAT] : "2018/09/05"[PDAT])
```
