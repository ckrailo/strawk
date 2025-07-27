BEGIN {
  print 1+2, 2-3, 4*5, 6/7, 8%9, 10^11
  print "----"
  a = 1
  a += 1
  print a
  a -= 1
  print a
  print ++a
  print a++
  print a
  a *= 2
  print a
  a /= 2
  print a
  a %= 3
  print a
  a ^= 2
  print a
}
