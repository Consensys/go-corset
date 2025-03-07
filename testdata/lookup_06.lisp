(module m1)
(defcolumns (X :i16))

(module m2)
(defcolumns Y)
;;
(deflookup test (Y) (m1.X))
