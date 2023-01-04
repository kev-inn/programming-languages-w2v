using System;
using System.Data.SQLite;

namespace DbMerger;

public static class Program
{
    public static void Main(string[] args)
    {
        if (args.Length == 0 || 
            args.Contains("-h"))
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("-h Help");
            Console.WriteLine("-s Source database file path");
            Console.WriteLine("-d Target database file path");
        }

        string? source = null;
        string? destination = null;

        if (args.Contains("-s"))
        {
            source = args[Array.IndexOf(args, "-s") + 1];
            
            // Check if source is a valid path to a file
            var sourceFile = new FileInfo(source);
            
            if (!sourceFile.Exists)
            {
                Console.WriteLine("Source file does not exist");
                return;
            }
        }

        if (args.Contains("-d"))
        {
            destination = args[Array.IndexOf(args, "-d") + 1];
            
            var destinationFile = new FileInfo(destination);
            
            if (!destinationFile.Exists)
            {
                Console.WriteLine("Destination file does not exist. Do you wish to create it? (y/n)");
                if (!destinationFile.Exists)
                {
                    // Create a new sqlite database file
                    SQLiteConnection.CreateFile(destination);
                    
                    // Open the connection
                    using var connection = new SQLiteConnection($"Data Source={destination};Version=3;");
                    connection.Open();
                    
                    // Execute some sql
                    using var createTableCode = new SQLiteCommand(connection);
                    createTableCode.CommandText = "CREATE TABLE IF NOT EXISTS \"code\" (\n\t\"id\"\tINTEGER,\n\t\"language\"\tTEXT NOT NULL,\n\t\"url\"\tTEXT NOT NULL,\n\t\"content\"\tTEXT NOT NULL,\n\t\"hash\"\tTEXT NOT NULL UNIQUE,\n\t\"size\"\tINTEGER NOT NULL DEFAULT 0,\n\tPRIMARY KEY(\"id\" AUTOINCREMENT)\n);";
                    
                    createTableCode.ExecuteNonQuery();

                    using var createTableProgress = new SQLiteCommand(connection);
                    createTableProgress.CommandText = "CREATE TABLE IF NOT EXISTS \"progress\" (\n    \t\"language\"\tTEXT NOT NULL,\n    \t\"query\"\tTEXT NOT NULL,\n    \t\"last_page\"\tINTEGER NOT NULL DEFAULT 0,\n    \tPRIMARY KEY(\"language\", \"query\")\n);";

                    createTableProgress.ExecuteNonQuery();
                    
                    connection.Close();
                }
                
                // var answer = Console.ReadLine();
                //
                // if (answer == "y")
                // {
                //     destinationFile.Create();
                // }
                // else
                // {
                //     Console.WriteLine("Aborting.");
                //     return;
                // }
            }
        }


        if (string.IsNullOrWhiteSpace(source) ||
            string.IsNullOrWhiteSpace(destination))
        {
            Console.WriteLine("Please provide source and destination file paths");
        }


        var foo = new SQLiteDbMerger();
        foo.Merge(source, destination);



        return;
    }
}

