BEGIN {
  print "ab" ~ /a/
  print "ab" ~ /c/
  print "ab" !~ /c/
}
