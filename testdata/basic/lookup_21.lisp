;; fragmented lookup with two two column targets
(deflookup test ((t1.Y1 t1.Y2) (t2.Z1 t2.Z2)) (s.X1 s.X2))
(module s)
(defcolumns (X1 :i16) (X2 :i16))

(module t1)
(defcolumns (Y1 :i16) (Y2 :i16))

(module t2)
(defcolumns (Z1 :i16) (Z2 :i16))
