(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 0 1))))
;; X == Y - 0
(defconstraint c2 () (vanishes! (- X Y (* 1 0))))
;; X == Y - 0
(defconstraint c3 () (vanishes! (- X Y (* 0 2))))
;; X == Y - 0
(defconstraint c4 () (vanishes! (- X Y (* 2 0))))
;; X == Y - 0
(defconstraint c5 () (vanishes! (- X Y (* 0 0 1))))
;; X == Y - 0
(defconstraint c6 () (vanishes! (- X Y (* 0 1 0))))
;; X == Y - 0
(defconstraint c7 () (vanishes! (- X Y (* 0 1 1))))
;; X == Y - 0
(defconstraint c8 () (vanishes! (- X Y (* 1 0 0))))
;; X == Y - 0
(defconstraint c9 () (vanishes! (- X Y (* 1 0 1))))
;; X == Y - 0
(defconstraint c10 () (vanishes! (- X Y (* 1 1 0))))
