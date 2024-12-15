;;error:2:1-2:blah
(module m1)
(defcolumns (X :i16@prove))
(defpermutation (Z) ((+ m1.X)))
