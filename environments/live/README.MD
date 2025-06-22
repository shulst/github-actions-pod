SQL om ervoor te zorgen dat we db's kunnen importeren:
```
EXEC sp_configure 'contained database authentication', 1;
GO
RECONFIGURE;
GO

EXEC sp_configure 'contained database authentication';
```
