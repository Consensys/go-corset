;; fragmented lookup with two targets
(deflookup test ((t1.Y) (t2.Z)) (s.X))
;;
(module t1)
(defcolumns (Y :i16))

(module t2)
(defcolumns (Z :i16))

(module s)
(defcolumns (X :i16))
