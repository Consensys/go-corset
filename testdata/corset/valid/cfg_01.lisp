(module f)
(defcolumns (x :i256) (y :i256) (res :i2))

(defcomputedcolumn (c1 :u1 :padding 1) (if (== 0 x) 1 0))
(defcomputedcolumn (c2 :u1 :padding 1) (if (== 0 y) 1 0))
(defcomputedcolumn (c3 :u1) (if (== 1 y) 1 0))

(defconstraint case_1 () (if (== 1 c1) (== res 0)))
(defconstraint case_2 () (if (∧ (== 0 c1) (== 1 c2)) (== res 1)))
(defconstraint case_3 () (if (∧ (== 0 c1) (== 1 c3)) (== res 2)))

;; (defconstraint case_4 ()
;;   (if (== 0 (+ c1 c2 c3)) (== res 3)))

(defconstraint case_4b ()
  (if (== 0 (+ c1 c2 c3)) (== res 3)))
