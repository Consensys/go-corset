(defpurefun (not! x) (- 1 (~ x)))
(defpurefun ((not_eq! :i16) x y) (not! (- x y)))
;;
(defcolumns (X :i16))
;; X != 1
(defconstraint X_t1 ()
  (for j [2] (for i [1] (not_eq! X i))))

(defconstraint X_t2 ()
  (for i [1] (for j [2] (not_eq! X i))))
