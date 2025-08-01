# hello
/([^,]+),([^,\n]+)*/ {
  for (g in $MATCHES) {
    print g, $MATCHES[g]
  }
}

