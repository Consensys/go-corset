(defcolumns (X :i16) (Y :i16))
;; X == Y - 0
(defconstraint c1 () (== 0 (- X Y (* 0 1))))
;; X == Y - 0
(defconstraint c2 () (== 0 (- X Y (* 1 0))))
;; X == Y - 0
(defconstraint c3 () (== 0 (- X Y (* 0 2))))
;; X == Y - 0
(defconstraint c4 () (== 0 (- X Y (* 2 0))))
;; X == Y - 0
(defconstraint c5 () (== 0 (- X Y (* 0 0 1))))
;; X == Y - 0
(defconstraint c6 () (== 0 (- X Y (* 0 1 0))))
;; X == Y - 0
(defconstraint c7 () (== 0 (- X Y (* 0 1 1))))
;; X == Y - 0
(defconstraint c8 () (== 0 (- X Y (* 1 0 0))))
;; X == Y - 0
(defconstraint c9 () (== 0 (- X Y (* 1 0 1))))
;; X == Y - 0
(defconstraint c10 () (== 0 (- X Y (* 1 1 0))))
