(defpurefun ((not_eq! :bool) x y) (!= x y))
;;
(defcolumns (X :i16))
;; X != 2 && X != 4 && X != 8
(defconstraint X_t1 ()
  (for i [1:3] (not_eq! X (^ 2 i))))
