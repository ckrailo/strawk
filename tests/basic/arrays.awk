BEGIN {
  a[1] = "a"
  print a[1]
  print a[2]
  print 1 in a
  print 2 in a
  delete a[1]
  print 1 in a
  a[2] = 1
  print ++a[2]
  print a[2]++
  print a[2]
}
