BEGIN {
  paragraph=0
}
# This regex matches each sentence in the input, and then prints the sentence. It works across newlines!
/(?s)(.*?)\n\n/ {
  paragraph += 1
  if paragraph % 2 == 0 {
    print $0
  }
}
