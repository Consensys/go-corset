(defpurefun (not! x) (- 1 (~ x)))
(defpurefun ((not_eq! :i16) x y) (not! (- x y)))
;;
(defcolumns (X :i16))
;; X != 1
(defconstraint X_t1 ()
  (for i [1] (not_eq! X i)))
