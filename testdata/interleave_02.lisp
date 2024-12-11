(defpurefun ((vanishes! :@loob) x) x)
;;
(defcolumns X Y)
(definterleaved Z (X Y))
;; Z[k]+1 == Z[k+1] || Z[k] == Z[k+1]
(defconstraint c1 ()
  (vanishes!
   (* (- (+ 1 Z) (shift Z 1)) (- Z (shift Z 1)))))
