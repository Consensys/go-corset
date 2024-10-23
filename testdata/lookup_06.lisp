(module m1)
(defcolumns X)

(module m2)
(defcolumns Y)
;;
(deflookup test (Y) (m1.X))
