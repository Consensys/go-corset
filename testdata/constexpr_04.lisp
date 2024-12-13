(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X Y)
;; X + 2 == Y + 2
(defconstraint c1 () (vanishes! (- (+ X 2) (+ Y (^ 2 1)))))
;; X + 4 == Y + 4
(defconstraint c1 () (vanishes! (- (+ X 4) (+ Y (^ 2 2)))))
;; X + 8 == Y + 8
(defconstraint c1 () (vanishes! (- (+ X 8) (+ Y (^ 2 3)))))
;; X + 16 == Y + 16
(defconstraint c1 () (vanishes! (- (+ X 16) (+ Y (^ 2 4)))))
;; X + 32 == Y + 32
(defconstraint c1 () (vanishes! (- (+ X 32) (+ Y (^ 2 5)))))
;; X + 64 == Y + 64
(defconstraint c1 () (vanishes! (- (+ X 64) (+ Y (^ 2 6)))))
;; X + 128 == Y + 128
(defconstraint c1 () (vanishes! (- (+ X 128) (+ Y (^ 2 7)))))
;; X + 256 == Y + 256
(defconstraint c1 () (vanishes! (- (+ X 256) (+ Y (^ 2 8)))))
