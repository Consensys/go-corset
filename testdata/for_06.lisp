(defpurefun ((not_eq! :bool) x y) (!= x y))
(defpurefun (f x y z) (for i [1] (not_eq! x i)))
;;
(defcolumns (X :i16))
;; X != 1
(defconstraint X_t1 () (f X X X))
