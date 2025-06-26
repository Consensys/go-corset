(defcolumns (X1 :i16) (X2 :i16) (Y1 :i16) (Y2 :i16) (Z1 :i16) (Z2 :i16))
;; fragmented lookup with two two column targets
(deflookup test ((Y1 Y2) (Z1 Z2)) (X1 X2))
