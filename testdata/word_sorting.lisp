;; A simple example which demonstrates how a sorting constraint can be
;; implemented on a column of bytes.

;; Input column
(defcolumns (X :u16))

;; Generated columns
(defcolumns Delta) ;; implied u16
(defcolumns (Byte_0 :u8) (Byte_1 :u8))

;; Ensure Delta is a u16
(defconstraint delta_type () (- Delta (+ (* 256 Byte_1) Byte_0)))

;; Delta == X - X[i-1]
(defconstraint sort () (- Delta (- X (shift X -1))))
