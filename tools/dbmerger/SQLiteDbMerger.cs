using System.Collections;
using System.Data.SQLite;

namespace DbMerger;

public class SQLiteDbMerger
{
    public void Merge(string sourceDbPath, string targetDbPath)
    {
        using var sourceDb = new SQLiteConnection("Data Source=" + sourceDbPath);
        using var targetDb = new SQLiteConnection("Data Source=" + targetDbPath);
        
        sourceDb.Open();
        targetDb.Open();
        
        // Copy over all data from sourceDb to targetDb
        var tables = new[] { "code", "progress" };
        
        // Code
        using var cmd = new SQLiteCommand("SELECT * FROM code", sourceDb);
        using var reader = cmd.ExecuteReader();
        while (reader.Read())
        {
            
            var insertCmd = new SQLiteCommand("INSERT INTO code (language, url, content, hash, size) VALUES (?, ?, ?, ?, ?)", targetDb);
            for (int i = 1; i < reader.FieldCount; i++)
            {
                insertCmd.Parameters.AddWithValue("@" + i, reader[i]);
            }
            
            // Try executing the command
            try
            {
                insertCmd.ExecuteNonQuery();
            }
            catch (SQLiteException e)
            {
                // Ignore duplicate key errors
                if (e.ErrorCode != 19)
                {
                    throw;
                }
            }
        }
        
        // Progress
        using var cmd2 = new SQLiteCommand("SELECT * FROM progress", sourceDb);
        using var reader2 = cmd2.ExecuteReader();
        while (reader2.Read())
        {
            var insertOrReplaceCmd = new SQLiteCommand("INSERT OR REPLACE INTO progress (language, query, last_page) VALUES (?, ?, ?)", targetDb);
            for (int i = 0; i < reader2.FieldCount; i++)
            {
                insertOrReplaceCmd.Parameters.AddWithValue("@" + i, reader2[i]);
            }
            
            insertOrReplaceCmd.ExecuteNonQuery();
        }
    }
}