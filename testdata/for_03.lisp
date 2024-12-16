(defpurefun (not! x) (- 1 (~ x)))
(defpurefun ((not_eq! :@loob) x y) (not! (- x y)))
;;
(defcolumns X)
;; X != 1
(defconstraint X_t1 ()
  (for i [1] (not_eq! X i)))
