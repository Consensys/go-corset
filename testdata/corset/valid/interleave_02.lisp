;;
(defcolumns (X :i16) (Y :i16))
(definterleaved Z (X Y))
;; Z[k]+1 == Z[k+1] || Z[k] == Z[k+1]
(defconstraint c1 ()
  (== 0
   (* (- (+ 1 Z) (shift Z 1)) (- Z (shift Z 1)))))
