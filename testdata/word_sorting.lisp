;; A simple example which demonstrates how a sorting constraint can be
;; implemented on a column of bytes.

;; Input column
(column X :u16)

;; Generated columns
(column Delta) ;; implied u16
(column Byte_0 :u8)
(column Byte_1 :u8)

;; Ensure Delta is a u16
(vanish delta_type (- Delta (+ (* 256 Byte_1) Byte_0)))

;; Delta == X - X[i-1]
(vanish sort (- Delta (- X (shift X -1))))
