BEGIN {
  a="123abc456xyz"
  b[1] = 1
  b[2] = 2
  b[3] = 3
  print a, length(a)
  print length(b)
  print sub(/abc/, "789", a)
  c="123abc456xyzabc"
  print gsub(/abc/, "789", c)
}
