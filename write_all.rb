codes = Dir["./codes/*"].sort 

File.open("./all_codes.md","w"){|f|

    codes.each{|filename|
        codename = filename.gsub("./codes/","").gsub(".md","")
        puts "code(\"#{codename}\")"
        file = File.open(filename)
        content = file.read.strip

        content = content.gsub("\n\n", "<><>").gsub("\n"," ").gsub("<><>","\n").gsub("  "," ")

        f.puts content        

        file.close
                
    }
}