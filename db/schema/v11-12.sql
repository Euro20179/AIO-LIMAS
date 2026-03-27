UPDATE entryInfo set recommendedby = concat('["', (select recommendedBy from entryInfo as e where entryInfo.itemId = e.itemId), '"]') where recommendedBy != '';
UPDATE entryInfo set recommendedby = '[]'  where recommendedBy = '';
UPDATE DBInfo SET version = 12;
