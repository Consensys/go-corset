(defpurefun (not! x) (- 1 (~ x)))
(defpurefun ((not_eq! :@loob) x y) (not! (- x y)))
;;
(defcolumns X)
;; X != 2 && X != 4 && X != 8
(defconstraint X_t1 ()
  (for i [1:3] (not_eq! X (^ 2 i))))
