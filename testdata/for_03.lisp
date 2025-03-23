(defpurefun ((not_eq! :bool) x y) (!= x y))
;;
(defcolumns (X :i16))
;; X != 1
(defconstraint X_t1 ()
  (for i [1] (not_eq! X i)))
