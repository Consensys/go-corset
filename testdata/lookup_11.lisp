(deflookup test (m1.Y) (m1.Z))
(module m1)
(defalias Z X)
(defcolumns (X :i16) (Y :i16))
