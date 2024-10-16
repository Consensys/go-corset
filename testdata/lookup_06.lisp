(module m1)
(defcolumns X)

(module m2)
(defcolumns Y)
;;
(lookup test (Y) (m1.X))
