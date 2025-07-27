BEGIN {
  a = 0
  while (a < 3) {
    print "a"
    a++
  }
  b = 0
  do {
    print "b"
  } while (b < 0)

  for (i = 0; i < 3; ++i) {
   print i 
  }

  for (i = 0; i < 3; ++i) {
   print "c"
   break
  }

  for (i = 0; i < 3; ++i) {
   print "d"
   continue
   print "e"
  }
}
