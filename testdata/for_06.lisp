(defpurefun (not! x) (- 1 (~ x)))
(defpurefun ((not_eq! :i16@loob) x y) (not! (- x y)))
(defpurefun (f x y z) (for i [1] (not_eq! x i)))
;;
(defcolumns X)
;; X != 1
(defconstraint X_t1 () (f X X X))
