# hello
BEGIN { 
  x=1
  y=1 
}

/ +/ {
  x += length($0)
}


/#+/ {
  print "rect", x, x+length($0), y, y+1
  x+=length($0)
}

/\n/ {
	x=1
	y++ 
}
