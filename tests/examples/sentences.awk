BEGIN {
  count=0
  words=0
}
# This regex matches each sentence in the input, and then prints the sentence. It works across newlines!
/(?s)(.*)\./ { 
  #gsub (Global SUBstitution replaces all newlines in the string with an empty string
  sentence = gsub(/\n/, "")
  sentence = sub(/^ /, "", sentence) #sub only replaces the first occurance of a regex
  print sentence
  splitwords = split(sentence, " ")
  count += 1
  words += length(splitwords)
}

END { print "Average length of sentence in text:", (words / count) }
