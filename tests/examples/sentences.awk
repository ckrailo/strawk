
/(?s)(.*)\./ { 
  sentence = gsub(/\n/, "")
  print sentence
}
