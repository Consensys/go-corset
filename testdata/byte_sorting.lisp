;; A simple example which demonstrates how a sorting constraint can be
;; implemented on a column of bytes.

;; Input column
(defcolumns (X :u8))

;; Generated column
(defcolumns (Delta :u8))

;; Delta == X - X[i-1]
(vanish sort (- Delta (- X (shift X -1))))
