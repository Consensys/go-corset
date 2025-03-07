(defpurefun (not! x) (- 1 (~ x)))
(defpurefun ((not_eq! :i16@loob) x y) (not! (- x y)))
;;
(defcolumns X)
;; X != 1 && X != 2 && X != 3
(defconstraint X_t1 ()
  (for i [1:3] (not_eq! X i)))
;; Syntactical variant
(defconstraint X_t2 ()
  (for i [1 :3] (not_eq! X i)))
;; Syntactical variant
(defconstraint X_t3 ()
  (for i [1: 3] (not_eq! X i)))
;; Syntactical variant
(defconstraint X_t4 ()
  (for i [ 1:3 ] (not_eq! X i)))
;; Syntactical variant
(defconstraint X_t5 ()
  (for i [ 1 :3 ] (not_eq! X i)))
;; Syntactical variant
(defconstraint X_t6 ()
  (for i [ 1: 3 ] (not_eq! X i)))
