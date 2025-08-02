BEGIN { sentence = 0}

/(?s)(.*)\./ { 
  sentence++
  s = sub(/\n/, "", $1)
  if sentence == 4 {
    new_sentence = ""
    for (char = 0; char < length(s); char++) {
      if char % 2  == 0 {
        c = toupper(substr(s, char, 1))
      } else {
        c = tolower(substr(s, char, 1))
      }
      new_sentence = new_sentence c
    }
    print new_sentence
  }
}
