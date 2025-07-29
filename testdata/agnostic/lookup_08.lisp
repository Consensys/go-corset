(module m1)
;;
(defcolumns (X :i128 :array [0:1]))
;;
(deflookup l1 (m2.Y) ((:: [X 1] [X 0])))

(module m2)
(defcolumns (Y :i256))
