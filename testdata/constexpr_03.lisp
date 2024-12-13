(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X Y)
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 0 1))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 1 0))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 0 2))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 2 0))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 0 0 1))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 0 1 0))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 0 1 1))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 1 0 0))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 1 0 1))))
;; X == Y - 0
(defconstraint c1 () (vanishes! (- X Y (* 1 1 0))))
